package main

import _ "embed"

// Project license text embedded from LICENSE at repo root
//go:embed LICENSE
var projectLicense string

// tviewLicense is the license text for github.com/rivo/tview (MIT License).
// Source: https://github.com/rivo/tview/blob/master/LICENSE
// The text below is the standard MIT license text.
const tviewLicense = `MIT License

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

func printLicenses() {
    // Application License
    println("=== LICENSE (Application) ===")
    if projectLicense != "" {
        println(projectLicense)
    } else {
        println("License file not embedded.")
    }

    // NOTICE for third-party dependencies
    println("\n=== NOTICE (Third-Party) ===")
    println("This product includes the following third-party software:\n")
    println("- github.com/rivo/tview (MIT License)")
    println("  See: https://github.com/rivo/tview")
    println("\n-- tview License (MIT) --\n" + tviewLicense)
}

