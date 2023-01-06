import os
from functools import partial
from typing import Dict, List, Optional, Tuple
from multiprocessing import Pool, current_process
import platform

from tqdm import tqdm
import pandas as pd
import numpy as np
import sqlite3
from py4j.java_gateway import JavaGateway, launch_gateway, GatewayParameters

STRING_LITERAL_TOKEN = "STRING_LITERAL"
INT_LITERAL_TOKEN = "INT_LITERAL"


def _generic_regex_tokenization(code: pd.Series):
    # TODO: maybe consider \t seperately for python and other such languages
    vars_or_keywords = r"\w+"
    dot_operator = r"\."
    # parantheses and other similar constructs
    parantheses_like = r"[<>/\\{}[\]()'\"]"
    # almost \W, but with some whitespaces. Captures rest of characters.
    non_words = r"[^a-zA-Z0-9_ \t\n\.<>/\\{}[\]()'\"]+"
    generic_regex = (
        rf"({vars_or_keywords}|{dot_operator}|{parantheses_like}|{non_words})"
    )

    return (
        code.str.lower()
        .str.replace(r"'(\\.|[^'\\])*'", f" {STRING_LITERAL_TOKEN} ", regex=True)
        .str.replace(r'"(\\.|[^"\\])*"', f" {STRING_LITERAL_TOKEN} ", regex=True)
        .str.replace(r"0x(\d|\w)+", f" {INT_LITERAL_TOKEN} ", regex=True)
        .str.replace(r"\d+", f" {INT_LITERAL_TOKEN} ", regex=True)
        .str.findall(generic_regex)
    )


if __name__ == "__main__":
    from pathlib import Path
    os.chdir(Path('..').absolute())

pool_size = 8
# "create_parser.(sh|bat)" script will create this
jarpath = (
    "./src/w2vtokenizer/target/w2vtokenizer-0.0.1-SNAPSHOT-jar-with-dependencies.jar"
)
classpath_seperator = ";" if platform.system() == "Windows" else ":"
classpath = classpath_seperator.join([jarpath])

gateway_port = launch_gateway(classpath=classpath, die_on_exit=True)
gateways = [
    JavaGateway(gateway_parameters=GatewayParameters(port=gateway_port))
    for _ in range(pool_size)
]


def code_tokenize_par(t, function_name):
    i, code = t
    res = getattr(
        gateways[i % pool_size].jvm.com.codetokenizer.Tokenizer, function_name
    )(code)
    return list(res)


def _antlr_tokenization(code: pd.Series, function_name: str):
    with Pool(pool_size) as p:
        tokenized_code = pd.Series(
            tqdm(
                p.imap(
                    partial(code_tokenize_par, function_name=function_name),
                    enumerate(code),
                ),
                total=len(code),
                smoothing=0.0001,
            ),
            index=code.index,
            dtype=object,
        )
    return tokenized_code


SPECIALIZED_TOKENIZATION = {
    "C++": partial(_antlr_tokenization, function_name="tokenizeCpp"),
    "C#": partial(_antlr_tokenization, function_name="tokenizeCsharp"),
    "Go": partial(_antlr_tokenization, function_name="tokenizeGo"),
    "Python": partial(_antlr_tokenization, function_name="tokenizePython3"),
}


def read_snippets_dataset(
    db_file_path: str, programming_language: Optional[str] = None
) -> pd.DataFrame:
    conn = sqlite3.connect(db_file_path)
    cur = conn.cursor()

    # check if database contains table "progress"
    foo = cur.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name='progress'"
    )
    # "our" database
    if len(list(foo)) == 0:
        if programming_language is None:
            snippets = cur.execute("SELECT language, snippet FROM snippets")
        else:
            snippets = cur.execute(
                f"SELECT language, snippet FROM snippets WHERE language='{programming_language}'"
            )
    else:
        if programming_language is None:
            snippets = cur.execute("SELECT language, content FROM code")
        else:
            snippets = cur.execute(
                f"SELECT language, content FROM code WHERE language='{programming_language}'"
            )

    return pd.DataFrame(snippets, columns=["language", "code"])


def read_lang_dataset(db_file_path: str) -> pd.DataFrame:
    conn = sqlite3.connect(db_file_path)
    cur = conn.cursor()
    data = cur.execute("SELECT language, content FROM code")
    data = pd.DataFrame(data, columns=["language", "code"])
    data.code = data.code.str.decode("utf-8", errors="replace")
    return data


def tokenize_dataset(dataset: pd.DataFrame):
    dataset = dataset.copy()
    for language in dataset.language.unique():
        code_selection = dataset.code[dataset.language == language]
        if language in SPECIALIZED_TOKENIZATION:
            dataset.code[dataset.language == language] = SPECIALIZED_TOKENIZATION[
                language
            ](code_selection)
        else:
            dataset.code[dataset.language == language] = _generic_regex_tokenization(
                code_selection
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


def main():
    ds = read_lang_dataset("data/codes_go.db")
    tokenize_dataset(ds)


if __name__ == "__main__":
    main()
