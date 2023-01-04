from typing import Dict, List, Optional, Tuple
from multiprocessing import Pool
import platform

from tqdm import tqdm
import pandas as pd
import numpy as np
import swifter
from swifter import set_defaults
import sqlite3
from antlr4 import *
from py4j.java_gateway import JavaGateway, launch_gateway, GatewayParameters

from .parsers.CPP14Lexer import CPP14Lexer
from .parsers.CPP14ParserListener import CPP14ParserListener
from .parsers.CPP14Parser import CPP14Parser
from .tokenizers import *

set_defaults(allow_dask_on_strings=True, progress_bar=True)


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


pool_size = 8
# "create_parser.(sh|bat)" script will create this
jarpath = "./src/java_parser/build/tokenizer.jar"
classpath_seperator = ";" if platform.system() == "Windows" else ":"
classpath = classpath_seperator.join(
    ["./src/java_parser/antlr-runtime-4.11.1.jar", jarpath]
)

gateway_port = launch_gateway(classpath=classpath, die_on_exit=True)
gateways = [
    JavaGateway(gateway_parameters=GatewayParameters(port=gateway_port))
    for _ in range(pool_size)
]


def code_tokenize_par(t):
    i, code = t
    res = gateways[i % pool_size].jvm.CppTokenizer.tokenize(code)
    return list(res)


def _cpp_tokenization(code: pd.Series):
    with Pool(pool_size) as p:
        tokenized_code = pd.Series(
            tqdm(
                p.imap(code_tokenize_par, enumerate(code)),
                total=len(code),
                smoothing=0.01,
            ),
            index=code.index,
            dtype=object,
        )
    return tokenized_code


SPECIALIZED_TOKENIZATION = {"C++": _cpp_tokenization}


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


def load(db_file_path: str, programming_language: Optional[str] = None):
    conn = sqlite3.connect(db_file_path)
    cur = conn.cursor()

    # check if database contains table "progress"
    foo = cur.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='progress'")
    # "our" database
    if len(list(foo)) == 0:
        if programming_language is None:
            snippets = cur.execute("SELECT language, snippet FROM snippets")
        else:
            snippets = cur.execute(f"SELECT language, snippet FROM snippets WHERE language='{programming_language}'")
    else:
        if programming_language is None:
            snippets = cur.execute("SELECT language, content FROM code")
        else:
            snippets = cur.execute(f"SELECT language, content FROM code WHERE language='{programming_language}'")

    return pd.DataFrame(snippets, columns=["language", "code"])


def tokenize_dataset(dataset: pd.DataFrame):
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
