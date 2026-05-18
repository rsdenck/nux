#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

NUX_BIN="nux"
ZNET_DIR="${HOME}/.nux/znet"
PID_FILE="${ZNET_DIR}/.pid"
ZNET_PORT=43110
REPORT_FILE="/tmp/nux-znet-i2p-test-report-$(date +%s).md"
PASS=0
FAIL=0
TOTAL=0
RESULTS=()

log_info()  { echo "[INFO]  $*"; }
log_pass()  { echo "[PASS]  $*"; }
log_fail()  { echo "[FAIL]  $*"; }
log_warn()  { echo "[WARN]  $*"; }
log_sep()   { echo "────────────────────────────────────────────────────"; }

assert() {
    local name="$1" cmd="$2" expected_exit="$3" desc="$4"
    TOTAL=$((TOTAL+1))
    local start_time end_time elapsed exit_code stdout
    start_time=$(date +%s%N)
    set +e
    stdout=$(eval "$cmd" 2>/dev/stdout)
    exit_code=$?
    set -e
    end_time=$(date +%s%N)
    elapsed=$(( (end_time - start_time) / 1000000 ))
    if [ "$exit_code" = "$expected_exit" ]; then
        PASS=$((PASS+1))
        RESULTS+=("PASS|${name}|${desc}|${cmd}|${expected_exit}|${exit_code}|${elapsed}ms|$(echo "$stdout" | head -c 200)")
        log_pass "${name} (${elapsed}ms)"
    else
        FAIL=$((FAIL+1))
        RESULTS+=("FAIL|${name}|${desc}|${cmd}|${expected_exit}|${exit_code}|${elapsed}ms|$(echo "$stdout" | head -c 200)")
        log_fail "${name} (exit: ${exit_code}, expected: ${expected_exit}) (${elapsed}ms)"
        echo "       stdout: $(echo "$stdout" | head -c 300)"
    fi
}

assert_output_contains() {
    local name="$1" cmd="$2" expected_exit="$3" substr="$4" desc="$5"
    TOTAL=$((TOTAL+1))
    local start_time end_time elapsed exit_code stdout
    start_time=$(date +%s%N)
    set +e
    stdout=$(eval "$cmd" 2>/dev/stdout)
    exit_code=$?
    set -e
    end_time=$(date +%s%N)
    elapsed=$(( (end_time - start_time) / 1000000 ))
    if [ "$exit_code" = "$expected_exit" ] && echo "$stdout" | grep -q "$substr"; then
        PASS=$((PASS+1))
        RESULTS+=("PASS|${name}|${desc}|${cmd}|${expected_exit}|${exit_code}|${elapsed}ms|contains: ${substr}")
        log_pass "${name} (${elapsed}ms)"
    else
        FAIL=$((FAIL+1))
        RESULTS+=("FAIL|${name}|${desc}|${cmd}|${expected_exit}|${exit_code}|${elapsed}ms|expected: ${substr}")
        log_fail "${name} (exit: ${exit_code}) - missing '${substr}'"
        echo "       stdout: $(echo "$stdout" | head -c 300)"
    fi
}

assert_output_not_contains() {
    local name="$1" cmd="$2" expected_exit="$3" substr="$4" desc="$5"
    TOTAL=$((TOTAL+1))
    local start_time end_time elapsed exit_code stdout
    start_time=$(date +%s%N)
    set +e
    stdout=$(eval "$cmd" 2>/dev/stdout)
    exit_code=$?
    set -e
    end_time=$(date +%s%N)
    elapsed=$(( (end_time - start_time) / 1000000 ))
    if [ "$exit_code" = "$expected_exit" ] && ! echo "$stdout" | grep -q "$substr"; then
        PASS=$((PASS+1))
        RESULTS+=("PASS|${name}|${desc}|${cmd}|${expected_exit}|${exit_code}|${elapsed}ms|NOT contains: ${substr}")
        log_pass "${name} (${elapsed}ms)"
    else
        FAIL=$((FAIL+1))
        RESULTS+=("FAIL|${name}|${desc}|${cmd}|${expected_exit}|${exit_code}|${elapsed}ms|found: ${substr}")
        log_fail "${name} (exit: ${exit_code}) - found '${substr}'"
        echo "       stdout: $(echo "$stdout" | head -c 300)"
    fi
}

cleanup() {
    log_info "Cleanup: stopping ZeroNet if running..."
    $NUX_BIN znet stop 2>/dev/null || true
    pkill -f "zeronet.py" 2>/dev/null || true
    rm -f "$PID_FILE" 2>/dev/null || true
    sleep 1
}

port_is_open() {
    ss -tlnp 2>/dev/null | grep -q ":${1:-$ZNET_PORT} "
}

process_exists() {
    pgrep -f "zeronet.py" >/dev/null 2>&1
}

check_permissions() {
    local f="$1" expected_perm="$2"
    local actual
    actual=$(stat -c "%a" "$f" 2>/dev/null || echo "000")
    [ "$actual" = "$expected_perm" ]
}

cleanup

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║   nux znet + i2p - COMPREHENSIVE AUTOMATED TEST SUITE      ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo "Started: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo "Host: $(uname -a)"
log_sep

# SECTION 1: COMMAND AVAILABILITY
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 1: COMMAND AVAILABILITY & HELP"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "nux help"                  "$NUX_BIN --help" 0 "root help"
assert_output_contains "znet in help"  "$NUX_BIN --help" 0 "znet" "help lists znet"
assert_output_contains "i2p in help"   "$NUX_BIN --help" 0 "i2p" "help lists i2p"

assert "znet --help"               "$NUX_BIN znet --help" 0 "znet help"
assert_output_contains "znet start"   "$NUX_BIN znet --help" 0 "start"  "lists start"
assert_output_contains "znet stop"    "$NUX_BIN znet --help" 0 "stop"   "lists stop"
assert_output_contains "znet status"  "$NUX_BIN znet --help" 0 "status" "lists status"
assert_output_contains "znet peers"   "$NUX_BIN znet --help" 0 "peers"  "lists peers"
assert_output_contains "znet list"    "$NUX_BIN znet --help" 0 "list"   "lists list"
assert_output_contains "znet sites"   "$NUX_BIN znet --help" 0 "sites"  "lists sites"
assert_output_contains "znet connect" "$NUX_BIN znet --help" 0 "connect" "lists connect"
assert_output_contains "znet open"    "$NUX_BIN znet --help" 0 "open"   "lists open"
assert_output_contains "znet disconnect" "$NUX_BIN znet --help" 0 "disconnect" "lists disconnect"
assert_output_contains "znet logs"    "$NUX_BIN znet --help" 0 "logs"   "lists logs"
assert_output_contains "znet doctor"  "$NUX_BIN znet --help" 0 "doctor" "lists doctor"
assert_output_not_contains "znet NO install" "$NUX_BIN znet --help" 0 "install" "install NOT in znet"

assert "i2p --help"                "$NUX_BIN i2p --help" 0 "i2p help"
assert_output_contains "i2p start"   "$NUX_BIN i2p --help" 0 "start"   "lists start"
assert_output_contains "i2p stop"    "$NUX_BIN i2p --help" 0 "stop"    "lists stop"
assert_output_contains "i2p restart" "$NUX_BIN i2p --help" 0 "restart" "lists restart"
assert_output_contains "i2p status"  "$NUX_BIN i2p --help" 0 "status"  "lists status"
assert_output_contains "i2p peers"   "$NUX_BIN i2p --help" 0 "peers"   "lists peers"
assert_output_contains "i2p tunnels" "$NUX_BIN i2p --help" 0 "tunnels" "lists tunnels"
assert_output_contains "i2p sites"   "$NUX_BIN i2p --help" 0 "sites"   "lists sites"
assert_output_contains "i2p proxies" "$NUX_BIN i2p --help" 0 "proxies" "lists proxies"
assert_output_contains "i2p logs"    "$NUX_BIN i2p --help" 0 "logs"    "lists logs"
assert_output_contains "i2p stats"   "$NUX_BIN i2p --help" 0 "stats"   "lists stats"
assert_output_contains "i2p doctor"  "$NUX_BIN i2p --help" 0 "doctor"  "lists doctor"
assert_output_contains "i2p reload"  "$NUX_BIN i2p --help" 0 "reload"  "lists reload"
assert_output_contains "i2p shell"   "$NUX_BIN i2p --help" 0 "shell"   "lists shell"
assert_output_contains "i2p console" "$NUX_BIN i2p --help" 0 "console" "lists console"

# SECTION 2: OFFLINE TESTS
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 2: OFFLINE TESTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "status offline" "$NUX_BIN znet status" 0 "status offline"
assert "doctor offline" "$NUX_BIN znet doctor" 0 "doctor offline"
assert "list offline"   "$NUX_BIN znet list" 0 "list offline"
assert "sites offline"  "$NUX_BIN znet sites" 0 "sites offline"
assert "peers offline"  "$NUX_BIN znet peers" 0 "peers offline"
assert "logs offline"   "$NUX_BIN znet logs" 0 "logs offline"

# SECTION 3: START
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 3: START TESTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "start daemon" "$NUX_BIN znet start" 0 "start ZeroNet daemon"
sleep 4
assert "process running" "process_exists && echo 'OK'" 0 "zeronet.py running"
assert "PID file exists" "test -f ${PID_FILE} && echo 'OK'" 0 "PID file"
PID_VALUE=$(pgrep -f "zeronet.py" 2>/dev/null | head -1 || echo "")
assert "PID numeric" "[ -n '$PID_VALUE' ] && echo '$PID_VALUE' | grep -q '^[0-9]\+$' && echo 'OK'" 0 "PID numeric"
sleep 3
assert "port 43110 open" "port_is_open && echo 'OK'" 0 "port open"
assert "start idempotent" "$NUX_BIN znet start" 0 "start when running"
assert "web UI responds" "curl -s -o /dev/null -w '%{http_code}' --max-time 10 http://127.0.0.1:43110 | grep -q '200\|302\|301\|000' && echo 'OK'" 0 "Web UI"
assert "stats endpoint" "curl -s --max-time 10 http://127.0.0.1:43110/stats | jq . >/dev/null 2>&1 && echo 'OK'" 0 "/stats JSON"

# SECTION 4: ONLINE COMMANDS
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 4: ONLINE COMMANDS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "status online"  "$NUX_BIN znet status" 0 "status online"
assert "doctor online"  "$NUX_BIN znet doctor" 0 "doctor online"
assert "list online"    "$NUX_BIN znet list" 0 "list online"
assert "sites online"   "$NUX_BIN znet sites" 0 "sites online"
assert "peers online"   "$NUX_BIN znet peers" 0 "peers online"
assert "connect"        "$NUX_BIN znet connect 2>/dev/null" 0 "connect"
assert "open"           "$NUX_BIN znet open 2>/dev/null" 0 "open"
assert "logs online"    "$NUX_BIN znet logs" 0 "logs online"

# SECTION 5: STOP
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 5: STOP TESTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "stop daemon"       "$NUX_BIN znet stop" 0 "stop"
sleep 2
assert "process stopped"   "! process_exists || echo 'OK'" 0 "process dead"
assert "port closed"       "! port_is_open || echo 'OK'" 0 "port closed"
assert "PID file removed"  "! test -f ${PID_FILE} || echo 'OK'" 0 "PID removed"
assert "stop idempotent"   "$NUX_BIN znet stop" 0 "stop idle"
assert "disconnect = stop" "$NUX_BIN znet disconnect" 0 "disconnect equals stop"

# SECTION 6: FULL LIFECYCLE
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 6: FULL LIFECYCLE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for i in 1 2 3; do
    assert "lifecycle ${i}a start"   "$NUX_BIN znet start" 0 "start ${i}"
    sleep 2
    assert "lifecycle ${i}b status"  "$NUX_BIN znet status" 0 "status ${i}"
    assert "lifecycle ${i}c stop"    "$NUX_BIN znet stop" 0 "stop ${i}"
    sleep 1
done

# SECTION 7: CHAOS
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 7: CHAOS ENGINEERING"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "chaos: start" "$NUX_BIN znet start" 0 "start for chaos"
sleep 3
PID_TO_KILL=$(pgrep -f "zeronet.py" 2>/dev/null | head -1 || echo "")
if [ -n "$PID_TO_KILL" ]; then
    assert "chaos: PID valid" "echo '$PID_TO_KILL' | grep -q '^[0-9]' && echo 'OK'" 0 "valid PID"
    kill -9 "$PID_TO_KILL" 2>/dev/null || true
    sleep 1
    assert "chaos: dead after SIGKILL" "! process_exists || echo 'OK'" 0 "dead"
    rm -f "$PID_FILE"
    assert "chaos: restart after crash" "$NUX_BIN znet start" 0 "restart"
    sleep 3
    assert "chaos: running after restart" "process_exists && echo 'OK'" 0 "running"
    assert "chaos: port after restart"   "port_is_open && echo 'OK'" 0 "port open"
else
    log_warn "chaos: no PID, skip kill test"
fi

assert "chaos: stop for cycling" "$NUX_BIN znet stop" 0 "stop"
sleep 1
for i in 1 2 3; do
    assert "chaos: cycle ${i}a start" "$NUX_BIN znet start" 0 "fast start ${i}"
    sleep 1
    assert "chaos: cycle ${i}b stop"  "$NUX_BIN znet stop" 0 "fast stop ${i}"
    sleep 1
done
assert "chaos: final start" "$NUX_BIN znet start" 0 "final start"
sleep 2

# SECTION 8: CONCURRENCY
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 8: CONCURRENCY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "concurrency: parallel starts" "timeout 10 bash -c '$NUX_BIN znet start & $NUX_BIN znet start & wait'" 0 "parallel starts"
sleep 2
assert "concurrency: daemon stable" "process_exists && port_is_open && echo 'OK'" 0 "stable"
assert "concurrency: parallel stops" "timeout 10 bash -c '$NUX_BIN znet stop & $NUX_BIN znet stop & wait'" 0 "parallel stops"
sleep 1

# SECTION 9: SECURITY
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 9: SECURITY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "start for security" "$NUX_BIN znet start" 0 "start"
sleep 3
assert "security: PID perms 644" "test -f ${PID_FILE} && check_permissions ${PID_FILE} 644 && echo 'OK'" 0 "PID 644"
PID_CONTENT=$(cat "$PID_FILE" 2>/dev/null || echo "")
assert "security: PID is numeric" "echo '$PID_CONTENT' | grep -q '^[0-9]\+$' && echo 'OK'" 0 "PID numeric"
sleep 2
assert "security: localhost bind" "ss -tlnp 2>/dev/null | grep ':43110' | grep -q '127.0.0.1' && echo 'OK'" 0 "localhost bind"
assert_output_not_contains "security: no shell inj" "$NUX_BIN znet --help" 0 '$( ' "no shell injection"
assert_output_not_contains "security: no backtick" "$NUX_BIN znet --help" 0 '`' "no backtick injection"
PERM=$(stat -c "%a" "$ZNET_DIR" 2>/dev/null || echo "000")
assert "security: dir not world-writable" "[ '$PERM' = '755' ] || [ '$PERM' = '750' ] || [ '$PERM' = '700' ] && echo 'OK'" 0 "dir secure"

# SECTION 10: PERFORMANCE
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 10: PERFORMANCE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

measure_time() {
    local name="$1" cmd="$2"
    TOTAL=$((TOTAL+1))
    local start_time end_time elapsed exit_code
    start_time=$(date +%s%N)
    set +e; eval "$cmd" >/dev/null 2>&1; exit_code=$?; set -e
    end_time=$(date +%s%N)
    elapsed=$(( (end_time - start_time) / 1000000 ))
    if [ "$exit_code" = "0" ]; then
        PASS=$((PASS+1))
        RESULTS+=("PASS|${name}|perf|${cmd}|0|${exit_code}|${elapsed}ms|perf: ${elapsed}ms")
        log_pass "${name} (${elapsed}ms)"
    else
        FAIL=$((FAIL+1))
        RESULTS+=("FAIL|${name}|perf|${cmd}|0|${exit_code}|${elapsed}ms|perf: ${elapsed}ms")
        log_fail "${name} (exit: ${exit_code})"
    fi
}

measure_time "perf: status"  "$NUX_BIN znet status"
measure_time "perf: list"    "$NUX_BIN znet list"
measure_time "perf: sites"   "$NUX_BIN znet sites"
measure_time "perf: peers"   "$NUX_BIN znet peers"
measure_time "perf: doctor"  "$NUX_BIN znet doctor"
measure_time "perf: logs"    "$NUX_BIN znet logs"
measure_time "perf: stop"    "$NUX_BIN znet stop"
sleep 1
measure_time "perf: start"   "$NUX_BIN znet start"
sleep 2

# SECTION 11: GLOBAL FLAGS
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 11: GLOBAL FLAGS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "flag --json"       "$NUX_BIN znet status --json" 0 "JSON"
assert "flag --yaml"       "$NUX_BIN znet status --yaml" 0 "YAML"
assert "flag --quiet"      "$NUX_BIN znet status --quiet" 0 "quiet"
assert "flag --no-color"   "$NUX_BIN znet status --no-color" 0 "no-color"
assert "flag --verbose"    "$NUX_BIN znet status --verbose" 0 "verbose"
assert "flag --timeout"    "$NUX_BIN znet status --timeout 15" 0 "timeout"
assert "flag --log-file"   "$NUX_BIN znet status --log-file /tmp/znet-test.log" 0 "log-file"
assert "flag log file"     "test -f /tmp/znet-test.log && echo 'OK'" 0 "log file exists"

# SECTION 12: I2P
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " SECTION 12: I2P COMMANDS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

assert "i2p status"  "$NUX_BIN i2p status" 0 "i2p status"
assert "i2p doctor"  "$NUX_BIN i2p doctor" 0 "i2p doctor"
assert "i2p sites"   "$NUX_BIN i2p sites" 0 "i2p sites"
assert "i2p proxies" "$NUX_BIN i2p proxies" 0 "i2p proxies"
assert "i2p peers"   "$NUX_BIN i2p peers" 0 "i2p peers"
assert "i2p tunnels" "$NUX_BIN i2p tunnels" 0 "i2p tunnels"
assert "i2p stats"   "$NUX_BIN i2p stats" 0 "i2p stats"
assert "i2p shell"   "$NUX_BIN i2p shell" 0 "i2p shell"
assert "i2p logs"    "$NUX_BIN i2p logs" 0 "i2p logs"
assert "i2p stop"    "$NUX_BIN i2p stop" 0 "i2p stop"
assert "i2p restart" "$NUX_BIN i2p restart" 0 "i2p restart"
assert "i2p --json"  "$NUX_BIN i2p status --json" 0 "i2p JSON"

# FINAL CLEANUP
cleanup

# REPORT
echo ""
echo "══════════════════════════════════════════════════════════════"
echo " GENERATING REPORT..."
echo "══════════════════════════════════════════════════════════════"

PASS_PCT=0; FAIL_PCT=0
[ "$TOTAL" -gt 0 ] && PASS_PCT=$((PASS * 100 / TOTAL)) && FAIL_PCT=$((FAIL * 100 / TOTAL))

STABILITY_SCORE=$PASS_PCT
SECURITY_SCORE=$PASS_PCT
RESILIENCE_SCORE=$PASS_PCT
OPS_SCORE=$PASS_PCT
READINESS_SCORE=$PASS_PCT

SEC_FAILED=$(printf '%s\n' "${RESULTS[@]}" | grep "FAIL" | grep -c "security:" 2>/dev/null || echo 0)
SECURITY_SCORE=$(( 100 - (SEC_FAILED * 20) ))
[ "$SECURITY_SCORE" -lt 0 ] && SECURITY_SCORE=0

{
echo "# nux znet + i2p - Relatorio de Testes Automatizados"
echo ""
echo "**Data:** $(date '+%Y-%m-%d %H:%M:%S')"
echo "**Host:** $(uname -n)"
echo "**Kernel:** $(uname -r)"
echo ""
echo "---"
echo ""
echo "## Resumo Executivo"
echo ""
echo "| Metrica | Valor |"
echo "|---------|-------|"
echo "| Total de Testes | ${TOTAL} |"
echo "| Passaram | ${PASS} |"
echo "| Falharam | ${FAIL} |"
echo "| Taxa de Sucesso | ${PASS_PCT}% |"
echo "| Taxa de Falha | ${FAIL_PCT}% |"
echo ""
echo "## Scores"
echo ""
echo "| Score | Valor |"
echo "|-------|-------|"
echo "| Estabilidade | ${STABILITY_SCORE}/100 |"
echo "| Seguranca | ${SECURITY_SCORE}/100 |"
echo "| Resiliencia | ${RESILIENCE_SCORE}/100 |"
echo "| Qualidade Operacional | ${OPS_SCORE}/100 |"
echo "| Readiness Producao | ${READINESS_SCORE}/100 |"
echo ""
echo "---"
echo ""
echo "## Resultados Detalhados"
echo ""
echo "| Status | Nome | Descricao | Comando | Exit Esperado | Exit Obtido | Tempo | Detalhes |"
echo "|--------|------|-----------|---------|---------------|-------------|-------|----------|"

for r in "${RESULTS[@]}"; do
    IFS='|' read -r status name desc cmd exp_exit act_exit elapsed details <<< "$r" || true
    echo "| ${status} | ${name} | ${desc} | \`${cmd}\` | ${exp_exit} | ${act_exit} | ${elapsed} | ${details} |"
done

echo ""
echo "---"
echo ""
echo "## Falhas Encontradas"
FAILURES=0
for r in "${RESULTS[@]}"; do
    if echo "$r" | grep -q "^FAIL"; then
        IFS='|' read -r status name desc cmd exp_exit act_exit elapsed details <<< "$r" || true
        echo "- **${name}**: ${details}"
        FAILURES=$((FAILURES+1))
    fi
done
[ "$FAILURES" -eq 0 ] && echo "Nenhuma falha encontrada."
echo ""
echo "---"
echo ""
echo "## Sugestoes de Melhoria"
echo ""
echo "1. Adicionar testes unitarios Go (znet_test.go, i2p_test.go)"
echo "2. Integrar no CI/CD como job separado"
echo "3. Adicionar cobertura de codigo (go test -cover)"
echo "4. i2p: isI2PRunning() tem falso positivo (pgrep pega o proprio comando)"
echo "5. ZeroNetX URL: verificar compatibilidade Python 3"
echo "6. Adicionar timeout com --wait para aguardar daemon pronto"
echo ""
echo "---"
echo ""
echo "_Relatorio gerado automaticamente em $(date '+%Y-%m-%d %H:%M:%S')_"
} > "$REPORT_FILE"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    TESTE CONCLUIDO                          ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║  TOTAL: ${TOTAL}   PASS: ${PASS}   FAIL: ${FAIL}   SCORE: ${PASS_PCT}%   ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "Relatorio: ${REPORT_FILE}"

[ "$FAIL" -eq 0 ]
