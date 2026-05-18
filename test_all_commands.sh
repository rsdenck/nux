#!/bin/bash
# Test all NUX commands

NUX="./nux"

echo "=== TESTE DO NUX CLI MASTER ==="
echo ""

# Test help
echo "1. Testing: nux --help"
$NUX --help 2>&1 | head -15
echo ""

# Test onboard
echo "2. Testing: nux onboard"
timeout 5 $NUX onboard 2>&1 | head -10
echo ""

# Test doctor
echo "3. Testing: nux doctor"
timeout 5 $NUX doctor 2>&1
echo ""

# Test disk usage
echo "4. Testing: nux disk usage"
timeout 5 $NUX disk usage 2>&1
echo ""

# Test network list
echo "5. Testing: nux network list"
timeout 5 $NUX network list 2>&1
echo ""

# Test service list
echo "6. Testing: nux service list"
timeout 5 $NUX service list 2>&1
echo ""

# Test process list
echo "7. Testing: nux process list"
timeout 5 $NUX process list 2>&1
echo ""

# Test bash exec
echo "8. Testing: nux bash exec 'echo test'"
timeout 5 $NUX bash exec "echo test" 2>&1
echo ""

# Test audit logins
echo "9. Testing: nux audit logins"
timeout 5 $NUX audit logins 2>&1 | head -10
echo ""

# Test pkg install
echo "10. Testing: nux pkg install test-pkg"
timeout 5 $NUX pkg install test-pkg 2>&1
echo ""

# Test skill list
echo "11. Testing: nux skill list"
timeout 5 $NUX skill list 2>&1 | head -15
echo ""

# Test agent status
echo "12. Testing: nux agent status"
timeout 5 $NUX agent status 2>&1
echo ""

echo "=== FIM DOS TESTES ==="
