#!/usr/bin/env python3
"""
This is a test file for syntax highlighting verification
Open this file in cooledit to verify Python syntax highlighting works
"""

import os
from typing import List, Optional

# Constants
MAX_SIZE = 100
VERSION = "1.0.0"
IS_ENABLED = True

# Class definition
class Person:
    """A class representing a person"""

    def __init__(self, name: str, age: int):
        self.name = name
        self.age = age

    def greet(self) -> str:
        return f"Hello, I'm {self.name}!"

# Function definition
def main():
    # Keywords: def, if, else, for, return, class, import
    name = "World"

    # Strings should be highlighted
    greeting = f"Hello, {name}!"
    print(greeting)

    # Numbers should be highlighted
    count = 42
    pi = 3.14159
    hex_num = 0xFF

    # Comments should be highlighted (like this one)

    # Operators
    if count > 10 and pi < 4.0:
        print("Math works!")

    # Built-in functions
    length = len(name)
    result = list(range(length))

    # List comprehension
    squares = [x**2 for x in range(10)]

    # Dictionary
    person = {"name": "Alice", "age": 30}

    # Lambda
    double = lambda x: x * 2

    return None

if __name__ == "__main__":
    main()
