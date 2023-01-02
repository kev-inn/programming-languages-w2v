from typing import Dict, List, Optional, Tuple

from tqdm import tqdm
import pandas as pd
import sqlite3


def _generic_regex():
    # TODO: maybe consider \t seperately for python and other such languages
    vars_or_keywords = r"\w+"
    dot_operator = r"\."
    # parantheses and other similar constructs
    parantheses_like = r"[<>/\\{}[\]()'\"]"
    # almost \W, but with some whitespaces. Captures rest of characters.
    non_words = r"[^a-zA-Z0-9_ \t\n\.<>/\\{}[\]()'\"]+"
    return rf"({vars_or_keywords}|{dot_operator}|{parantheses_like}|{non_words})"


GENERIC_REGEX = _generic_regex()


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


def tokenize_dataset(dataset: pd.DataFrame):
    dataset["code"] = (
        dataset["code"]
        .str.lower()
        .str.replace(r"'(\\.|[^'\\])*'", " STRING_LITERAL ", regex=True)
        .str.replace(r'"(\\.|[^"\\])*"', " STRING_LITERAL ", regex=True)
        .str.replace(r"0x(\d|\w)+", " HEXNUMBER ", regex=True)
        .str.replace(r"\d+", " NUMBER ", regex=True)
        .str.findall(GENERIC_REGEX)
    )
    return dataset


def get_vocab_mapping(
    whole_tokenized_dataset: pd.DataFrame,
) -> Tuple[Dict[str, int], List[str]]:
    # TODO: initialize with whole vocab
    words = set()
    whole_tokenized_dataset["code"].apply(words.update)
    words = sorted(words)
    int2word = words
    word2int = {w: i for i, w in enumerate(words)}
    return word2int, int2word
