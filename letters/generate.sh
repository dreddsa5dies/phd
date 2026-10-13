#!/usr/bin/env -S gawk -f

BEGIN {
    FS = "\t"
    pwd = ENVIRON["PWD"]; if (pwd == "") pwd = "."
    osf = "OSFONTDIR=\"" pwd "/fonts\" "
    system(osf "luaotfload-tool --update >/dev/null 2>&1")
}

NF > 0 {
    defs = NR "-defs.tex"
    printf "\\def\\toindex{%s}\n",       esc($1) > defs
    printf "\\def\\tocity{%s}\n",        esc($2) > defs
    printf "\\def\\toaddress{%s}\n",     esc($3) > defs
    printf "\\def\\toinstitution{%s}\n", esc($4) > defs
    printf "\\def\\towhom{%s}\n",        esc($5) > defs
    close(defs)

    cmd = osf "luatex -interaction=nonstopmode -halt-on-error -jobname " NR " letters.tex"
    if (system(cmd) != 0)
        print "luatex failed for job " NR > "/dev/stderr"
}

END {
    system("rm -f letters.pdf")
    system("pdftk [0-9]*.pdf cat output letters.pdf.backup 2>/dev/null")
    system("rm -f *.log [0-9]*.pdf [0-9]*-defs.tex")
    system("mv letters.pdf.backup letters.pdf")
}

function esc(s,   out, i, n, c) {
    out = ""
    n = length(s)
    for (i = 1; i <= n; i++) {
        c = substr(s, i, 1)
        if      (c == "\\") out = out "\\char92{}"
        else if (c == "{")  out = out "\\{"
        else if (c == "}")  out = out "\\}"
        else if (c == "#")  out = out "\\#"
        else if (c == "$")  out = out "\\$"
        else if (c == "%")  out = out "\\%"
        else if (c == "&")  out = out "\\&"
        else if (c == "_")  out = out "\\_"
        else if (c == "^")  out = out "\\char94{}"
        else if (c == "~")  out = out "\\char126{}"
        else                out = out c
    }
    return out
}