# Contributing to gnark-crypto

### Table of Contents

[Code of Conduct](#code-of-conduct)

[How to Contribute](#how-to-contribute)

* [Guidelines for Non-Code and other Trivial Contributions](#guidelines-for-non-code-and-other-trivial-contributions)
* [Reporting Bugs](#reporting-bugs)
* [Suggesting Enhancements](#suggesting-enhancements)
* [Pull Requests](#pull-requests)



## Code of Conduct
* This project is governed by the [gnark-crypto Code of Conduct](CODE_OF_CONDUCT.md). By participating, 
you are agreeing to uphold this code. Please report unacceptable behavior.
## How to Contribute

### Guidelines for Non-Code and other Trivial Contributions
Please keep in mind that we do not accept non-code contributions like fixing comments, typos or some other trivial fixes. Although we appreciate the extra help, managing lots of these small contributions is unfeasible, and puts extra pressure in our continuous delivery systems (running all tests, etc). Feel free to open an issue pointing to any of those errors, and we will batch them into a single change.

### Reporting Bugs
#### Before Submitting A Bug 
* Ensure the bug is not already reported by searching on GitHub under 
[Issues](https://github.com/consensys/gnark-crypto/issues).
#### How Do I Submit a (Good) Bug?
* If you are unable to find an open issue addressing the problem, open a new one. Be sure to include a 
**title and clear description**, as much relevant information as possible, and a **code sample** or 
an **executable test case** demonstrating the unexpected behavior.
* Describe the **exact steps** to **reproduce the problem** in as many details as possible. When 
listing steps, don't just say what you did, but explain how you did it. 
* Provide **specific examples** to demonstrate the steps. Include links to files or GitHub projects, or 
copy/pasteable snippets, which you use in those examples. If you're providing snippets in the issue, 
use [Markdown code blocks](https://help.github.com/articles/getting-started-with-writing-and-formatting-on-github/).
* Describe the **behavior you observed** after following the steps and explain the 
problem with that behavior.
* Explain the **behavior you expected** instead and why.
* **Can you reliably reproduce the issue?** If not, provide details about how often the problem 
happens and under which conditions it normally happens.

### Suggesting Enhancements
#### Before Submitting An Enhancement Suggestion
* [Search](https://github.com/consensys/gnark-crypto/issues) to see if the enhancement has already been 
suggested. If it has, add a comment to the existing issue instead of opening a new one.

#### How Do I Submit A (Good) Enhancement Suggestion?
Enhancement suggestions are tracked as GitHub issues. Create an issue on and provide 
the following information:

* Use a **clear and descriptive title** for the issue to identify the suggestion.
* Provide a **step-by-step description** of the suggested enhancement in as much detail as possible.
* Describe the **current behavior** and explain the **behavior you expect** instead and why.
* Explain why this enhancement would be useful to other users.
* Specify the **name and version of the OS** you're using.
* Specify the **name and version of any relevant packages**.

### Pull Requests
There are a number of automated checks:
* `go fmt`
* `go vet`

If these checks pass, pull requests will be reviewed by the project team against criteria including:
* purpose - is this change useful
* test coverage - are there unit/integration/acceptance tests demonstrating the change is effective
* code consistency - naming, comments, design
* changes that are solely formatting are likely to be rejected

Always write a clear log message for your commits. One-line messages are fine for small changes, but 
bigger changes should contain more detail.
