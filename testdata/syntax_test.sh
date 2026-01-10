#!/bin/bash
# This is a test file for syntax highlighting verification
# Open this file in cooledit to verify Bash syntax highlighting works

# Variables
NAME="World"
COUNT=42
IS_ENABLED=true

# Strings should be highlighted
echo "Hello, $NAME!"
echo 'Single quoted string'

# Numbers
PORT=8080
HEX_VAL=0xFF

# Comments should be highlighted (like this one)

# Function definition
greet() {
    local name=$1
    echo "Hello, $name!"
}

# Conditional
if [ $COUNT -gt 10 ]; then
    echo "Count is greater than 10"
elif [ $COUNT -eq 10 ]; then
    echo "Count is exactly 10"
else
    echo "Count is less than 10"
fi

# Loop
for i in 1 2 3 4 5; do
    echo "Number: $i"
done

# While loop
while [ $COUNT -gt 0 ]; do
    COUNT=$((COUNT - 1))
done

# Case statement
case $NAME in
    "World")
        echo "Default name"
        ;;
    *)
        echo "Custom name"
        ;;
esac

# Command substitution
DATE=$(date +%Y-%m-%d)
FILES=`ls -la`

# Operators and redirects
cat file.txt > output.txt 2>&1
grep "pattern" file.txt | wc -l

# Exit code
exit 0
