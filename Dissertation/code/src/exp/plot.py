#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import argparse
import os
import sys
from math import pi

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd


def read_csv_safe(filepath, delimiter, **kwargs):
    if not os.path.exists(filepath):
        raise FileNotFoundError(f"Файл не найден: {filepath}")
    return pd.read_csv(filepath, delimiter=delimiter, encoding="utf-8-sig", **kwargs)


def clean_percentage_column(df, column_name):
    if column_name in df.columns:
        df[column_name] = (
            df[column_name].astype(str).str.replace("%", "", regex=False).astype(float)
        )
    return df


def normalize_by_scenario(df_scenario, metrics):
    """Нормализует метрики внутри сценария: 1 - всегда лучший результат."""
    df_norm = df_scenario.copy()
    for metric, better_is_higher in metrics.items():
        values = pd.to_numeric(df_norm[metric], errors="coerce")
        if values.isnull().any():
            values = values.fillna(0)
        min_val, max_val = values.min(), values.max()
        if max_val == min_val:
            norm = [0.5] * len(values)
        else:
            if better_is_higher:
                norm = (values - min_val) / (max_val - min_val)
            else:
                norm = (max_val - values) / (max_val - min_val)  # инверсия
        df_norm[f"norm_{metric}"] = norm
    return df_norm


def plot_radar_for_scenario(
    df_scenario_norm, metrics, metric_labels, scenario_name, output_dir
):
    angles = np.linspace(0, 2 * pi, len(metrics), endpoint=False).tolist()
    angles += angles[:1]

    fig, ax = plt.subplots(figsize=(8, 8), subplot_kw={"projection": "polar"})
    ax.set_theta_offset(pi / 2)
    ax.set_theta_direction(-1)

    colors = plt.cm.tab10(np.linspace(0, 1, len(df_scenario_norm)))
    line_styles = ["-", "--", "-.", ":", (0, (3, 1, 1, 1)), (0, (5, 1))]
    markers = ["o", "s", "^", "D", "v", "<", ">", "p", "*", "h"]

    for idx, (_, row) in enumerate(df_scenario_norm.iterrows()):
        strategy = row["Стратегия"]
        values = [row[f"norm_{m}"] for m in metrics.keys()] + [
            row[f"norm_{m}"] for m in metrics.keys()
        ][:1]
        color = colors[idx % len(colors)]
        linestyle = line_styles[idx % len(line_styles)]
        marker = markers[idx % len(markers)]

        ax.plot(
            angles,
            values,
            linewidth=2.5,
            linestyle=linestyle,
            color=color,
            marker=marker,
            markersize=6,
            label=strategy,
        )
        ax.fill(angles, values, alpha=0.08, color=color)

    ax.set_xticks(angles[:-1])
    labels = [
        f"{metric_labels[m]}¹" if not metrics[m] else metric_labels[m]
        for m in metrics.keys()
    ]
    ax.set_xticklabels(labels, fontsize=9)
    ax.set_ylim(0, 1)
    ax.set_yticks([0.2, 0.4, 0.6, 0.8, 1.0])
    ax.set_yticklabels(["0.2", "0.4", "0.6", "0.8", "1.0"], fontsize=8)
    ax.grid(True, linestyle=":", alpha=0.5)
    ax.set_title(
        f"Сценарий: {scenario_name}\n(¹ - инвертировано: большее значение = лучше)",
        fontsize=12,
        fontweight="bold",
        pad=20,
    )
    ax.legend(
        loc="upper left",
        bbox_to_anchor=(1.2, 1.0),
        fontsize=9,
        frameon=True,
        fancybox=True,
        shadow=True,
    )

    output_path = os.path.join(output_dir, f"radar_{scenario_name}.png")
    plt.tight_layout()
    plt.savefig(output_path, dpi=300, bbox_inches="tight")
    plt.close()
    print(f"Сохранено: {output_path}")


def main():
    parser = argparse.ArgumentParser(
        description='Радарные диаграммы с унификацией направления "чем больше, тем лучше"'
    )
    parser.add_argument("input_dir", nargs="?", help="Путь к папке с CSV-файлами")
    parser.add_argument(
        "--input-dir",
        "-d",
        dest="input_dir_opt",
        help="Альтернативный способ указать папку",
    )
    args = parser.parse_args()
    work_dir = args.input_dir or args.input_dir_opt
    if not work_dir:
        print("Ошибка: не указана папка с данными.")
        sys.exit(1)
    if not os.path.isdir(work_dir):
        print(f"Ошибка: '{work_dir}' не является папкой")
        sys.exit(1)

    gini_path = os.path.join(work_dir, "gini_results.csv")
    summary_path = os.path.join(work_dir, "summary_results.csv")
    try:
        gini_df = read_csv_safe(gini_path, delimiter=";")
        summary_df = read_csv_safe(summary_path, delimiter=",")
        gini_df.columns = gini_df.columns.str.strip()
        summary_df.columns = summary_df.columns.str.strip()
        summary_df = clean_percentage_column(summary_df, "% выполнения")

        merged = pd.merge(
            gini_df, summary_df, on=["Сценарий", "Стратегия"], how="inner"
        )
        merged = merged[
            [
                "Сценарий",
                "Стратегия",
                "Среднее значение (задачи)",
                "Среднее значение (энергия)",
                "% выполнения",
                "Ср. энергия на задачу",
                "Общее время (µs)",
            ]
        ]
        merged.columns = [
            "Сценарий",
            "Стратегия",
            "gini_tasks",
            "gini_energy",
            "pct_completed",
            "avg_energy_per_task",
            "time_us",
        ]
        for col in [
            "gini_tasks",
            "gini_energy",
            "pct_completed",
            "avg_energy_per_task",
            "time_us",
        ]:
            merged[col] = pd.to_numeric(merged[col], errors="coerce")

        metrics = {
            "gini_tasks": False,
            "gini_energy": False,
            "pct_completed": True,
            "avg_energy_per_task": False,
            "time_us": False,
        }
        metric_labels = {
            "gini_tasks": "Джини (задачи)",
            "gini_energy": "Джини (энергия)",
            "pct_completed": "Выполнено задач, %",
            "avg_energy_per_task": "Ср. энергия / задача",
            "time_us": "Время, мкс",
        }
    except Exception as e:
        print(f"Ошибка обработки данных: {e}")
        sys.exit(1)

    for scenario in merged["Сценарий"].unique():
        df_scen = merged[merged["Сценарий"] == scenario].copy()
        df_scen_norm = normalize_by_scenario(df_scen, metrics)
        plot_radar_for_scenario(
            df_scen_norm, metrics, metric_labels, scenario, work_dir
        )

    print("Готово.")


if __name__ == "__main__":
    main()
