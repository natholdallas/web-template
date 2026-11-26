# Rewrite the production-relevant keys of a deployed conf.toml.
#
# Called by main.sh patch_conf_toml on the *remote* copy (fetched into the
# packaging dir first), so a deploy is authoritative about:
#
#   [app]       debug = false, swagger = false   (never ship the dev CORS list or
#                                                 the public docs route)
#   [app]       nginx = true                     (only when web_server != none)
#   [resources] media-url = "https://<media>"    (only when domain_media is set)
#   [cors]      prd = [...]                      (rebuild: keep non-URL schemes,
#                                                 ensure the Tauri origins, add
#                                                 http+https for each site domain)
#
# Anything unexpected in the structure is a hard failure: guessing here would
# silently deploy the wrong CORS list or a debug backend.
#
# Invoked as: awk -v media_url=... -v sites="a b" -v nginx_mode=true|false \
#                  -f assets/patch-conf.awk conf.toml > conf.toml.patched

function trim(s) {
    sub(/^[ \t]+/, "", s)
    sub(/[ \t]+$/, "", s)
    return s
}

function die(msg) {
    printf "patch-conf: %s\n", msg > "/dev/stderr"
    exit 2
}

function keep(v) {
    if (v == "" || (v in seen)) return
    seen[v] = 1
    order[++on] = v
}

function add_site(s) {
    keep("http://" s)
    keep("https://" s)
}

# Emit the rebuilt prd array. Existing entries are already collected (filtered);
# the Tauri webview origins are always present, then every current site domain.
function flush_prd(   i, out) {
    if (sect != "cors" || prd_done) return
    prd_done = 1
    for (i = 1; i <= 3; i++) keep(tauri[i])
    for (i = 1; i <= nsites; i++) if (sites_arr[i] != "") add_site(sites_arr[i])
    out = ""
    for (i = 1; i <= on; i++) out = out (out == "" ? "" : ", ") "\"" order[i] "\""
    print "prd = [" out "]"
}

function leave_section() {
    if (sect == "cors") {
        if (!prd_seen) flush_prd()
    } else if (sect == "resources" && media_url != "" && !media_seen) {
        print "media-url = \"" media_url "\" # auto (main.sh)"
        media_seen = 1
    }
}

BEGIN {
    split("tauri://localhost http://tauri.localhost https://tauri.localhost", tauri, " ")
    nsites = split(sites, sites_arr, " ")
    sect = ""
    prd_seen = 0
    prd_done = 0
    media_seen = 0
    debug_seen = 0
    swagger_seen = 0
    nginx_seen = 0
    cors_seen = 0
    resources_seen = 0
    in_prd = 0
}

{
    line = $0
    t = trim(line)
    if (t ~ /^\[/) {
        leave_section()
        sect = t
        gsub(/[][[:space:]]/, "", sect)
        print line
        if (sect == "resources") resources_seen = 1
        if (sect == "cors") {
            cors_seen = 1
            prd_seen = 0
            prd_done = 0
        }
        next
    }
    if (sect == "cors") {
        if (in_prd) {
            collect = t
            gsub(/[]["]/, "", collect)
            n = split(collect, parts, ",")
            for (i = 1; i <= n; i++) {
                v = trim(parts[i])
                if (v != "" && v !~ /^https?:\/\//) keep(v)
            }
            if (t ~ /\]/) in_prd = 0
            prd_seen = 1
            if (!in_prd) flush_prd()
            next
        }
        if (t ~ /^prd[ \t]*=/) {
            prd_seen = 1
            collect = t
            sub(/^prd[ \t]*=[ \t]*\[/, "", collect)
            gsub(/[]["]/, "", collect)
            n = split(collect, parts, ",")
            for (i = 1; i <= n; i++) {
                v = trim(parts[i])
                if (v != "" && v !~ /^https?:\/\//) keep(v)
            }
            if (t ~ /\]/) flush_prd()
            else in_prd = 1
            next
        }
        print line
        next
    }
    if (sect == "app") {
        if (t ~ /^debug[ \t]*=/) {
            print "debug = false"
            debug_seen = 1
            next
        }
        if (t ~ /^swagger[ \t]*=/) {
            print "swagger = false"
            swagger_seen = 1
            next
        }
        if (nginx_mode == "true" && t ~ /^nginx[ \t]*=/) {
            print "nginx = true"
            nginx_seen = 1
            next
        }
        if (t ~ /^nginx[ \t]*=/) nginx_seen = 1
        print line
        next
    }
    if (sect == "resources" && media_url != "" && t ~ /^media-url[ \t]*=/) {
        print "media-url = \"" media_url "\" # auto (main.sh)"
        media_seen = 1
        next
    }
    print line
}

END {
    leave_section()
    if (sect == "cors" && !prd_done) flush_prd()
    if (!debug_seen) die("conf.toml has no [app] debug key")
    if (!swagger_seen) die("conf.toml has no [app] swagger key")
    if (nginx_mode == "true" && !nginx_seen) die("conf.toml has no [app] nginx key")
    if (!cors_seen) die("conf.toml has no [cors] section")
    if (media_url != "" && !media_seen && !resources_seen) die("conf.toml has no [resources] section")
}
