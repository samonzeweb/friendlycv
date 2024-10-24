# friendlycv

## Description

**friendlycv** stands for *friendly condition variable*.

It is a replacement for the `sync.Cond` condition variable implementation in the Go standard library, accepting a context into the `Wait` method, and so allowing proper timeouts and cancellations management.

This package uses a simple implementation based on channels, and can't compete with `sync.Cond` from a performance point of view.

## Licence

Released under the MIT License, see LICENSE.txt for more information.