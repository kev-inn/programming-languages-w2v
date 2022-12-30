from typing import Dict, List, Optional, Tuple
import re

import pandas as pd
import sqlite3


def read_dataset(
    db_file_path: str, programming_language: Optional[str] = None
) -> pd.DataFrame:
    conn = sqlite3.connect(db_file_path)
    cur = conn.cursor()

    if programming_language is None:
        snippets = cur.execute("SELECT language, snippet FROM snippets")
    else:
        snippets = cur.execute(
            f"SELECT language, snippet FROM snippets WHERE language='{programming_language}'"
        )
    return pd.DataFrame(snippets, columns=["language", "code"])


def _tokenize_row(row) -> pd.DataFrame:
    language, code = row

    # TODO: maybe consider \t seperately for python and other such languages
    vars_or_keywords = r"\b\w+\b"
    dot_operator = r"\."
    # parantheses and other similar constructs
    parantheses_like = r"[<>/\\{}[\]()'\"]"
    # almost \W, but with some whitespaces. Captures rest of characters, but non-greedy!
    non_words = r"\B[^a-zA-Z0-9_ \t]+?\B"
    tokens_regex = re.compile(
        rf"({vars_or_keywords}|{dot_operator}|{parantheses_like}|{non_words})"
    )
    tokens = tokens_regex.findall(code)
    return language, tokens


def tokenize_dataset(dataset: pd.DataFrame):
    return dataset.apply(_tokenize_row, "columns", result_type="broadcast")


def get_vocab_mapping(
    tokenized_dataset: pd.DataFrame,
) -> Tuple[Dict[str, int], List[str]]:
    # TODO: initialize with whole vocab
    int2word = None
    word2int = None  # {w: i for i, w in enumerate(words)}
    return word2int, int2word
