# API Documentation

## Introduction
    This document describes the use of the Lua API for declaring the
    behavior of a virtual plant. The set Lua API functions can be
    divided into blocking, game state changing functions, as well as
    non-blocking information gathering ones. These functions are
    written from the perspective of the 'tip' of a plant.

## Environment modifying functions

### grow("DIR")
    This function takes a string literal chosen from {"left", "right",
    "up", "down"} which instructs the tip to grow further. It has a
    price in Energy, and will block until the plant has a sufficient
    amount.

### split("DIR", "META")
    This function splits the tip of the current node into two: one new
    tip and itself. The new tip will have its own copy of the script
    you defined, starting at the top. The <DIR> argument represents
    the direction that the new node will begin to travel, and the
    <META> argument sets a variable accessible in the new tip with the
    meta() function. This function requires Energy to execute.

### terminate()
    This function turns the current active tip node into a passive
    plant node.

### spawn()
    This function calls terminates() the current tip, and releases a
    'spore', a game object that gently falls to the ground and plants
    a new instance of your plant. It requires Energy to execute.

### wait()
    This function pauses your tip until the next tick.

## Information gathering functions

### meta()
    This functions returns the string passed in from the parent split
    call. If called from the stem, it returns "".

### get_energy()
    This function return an integer representing your plant's current
    energy level.

### get_age()
    This function return an integer representing your plant's age.

### lighting("DIR")
    This function takes in a Direction String and returns an integer
    representing the light currently available at that node.
