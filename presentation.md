# Präsentation AIR

## Aufteilung: 
- HELLO (Tobi)
- Intro / Motivation (Tobi)
- Dataset and IR methods (Kevin)
- Results & Analysis (Phil)
- Conclusion (Manu)

## Potential questions: 
- Commercial usage. Imagine selling this to GitHub for example
- Improvements? 
  - How would you compare programming languages in a more meaningful way? (Other than a bigger dataset)
  - How is copilot designed? Why is it better than what we did?
- Other forms of analysis? 
  - Doc2Vec to guess what the code file is about (domain, application, query)
- Are there other things we can infer from textual/context similarities?


## 1. Introduction (Tobi):
- NERD origin story
- Which programming language is the best? 
- Settle the debate by using what we learned to textually analyze four different languages
- There's some common ground between
- Meaning of keywords that are shared between all languages (if, else, for, return)
- Differences in these keywords clarify specific usage of the language
- We expected a shitty analysis of semantics (i.e. )


## 2. Dataset and IR methods (Kevin):
- How we did it
- Unsupervised w2v is good for this
- Snippets-dev
  - Variable names are hard to get rid of
  - Solve this by using ANTLR4 (Grammar, Lexer, Parser)
- New dataset because ANTLR4 doesn't really work with small snippets
  - Goal was to get 2k files per language from GitHub (secondary rate limit der HS)
  - Filter out variables, floats, strings (best effort method)
  - Not perfect, but good enough for what we wanted to do
- Used FastText & W2V
  - Keywords prefixed with language-identifier: Python_if
- For analysis we used UMAP to create WordCloud plot

Used queries for datafetching: 
- GO: web, tool, framework, database, *, automation
- C#: automation, dotnet, web, game, template
- Python: automation, tool, auto, web
- C++: cache, example, class, embedded, posix, wrapper, client, tcp, database, automation, framework


## 3. Results & Analysis (Phil):
- Three languages (C++, GO, C#) are a cluster (strongly typed)
- Python is more on it's own (loosely typed)
- Vocabulary, Object oriented
- class seperatee learned from other langs
- catch c# -> try except
- syntax unabhängig -> try exept beieinandergo ; python \n, parser
- better result with larger datasets
- mehr code mit besserer bedeutung; C++ try catch with c# try catch
- average over all languages, c++-Go, C# etwas darüber, python wo ganz anders


## 4. Conclusion (Manu):
- We expected continue + break to be more similar. They're basically the same in all languages. 
- Bigger dataset of higher quality data would lead to better results, even with this method of analysis.
- Jupyter + git = not good
- Analysis of semantics / syntax similarities
- Usage domain of languages and their applications introduce bias
- Dataset was fetched with different queries ('game' for C#, 'embedded' for C++)
- count(if) >> count(for). so analyzing difference between them is unfair and biased