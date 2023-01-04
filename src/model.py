from torch import nn

EMBED_DIMENSION = 126
EMBED_MAX_NORM = 1


class SkipGramModel(nn.Module):
    """
    Implementation of Skip-Gram model described in paper:
    https://arxiv.org/abs/1301.3781
    """

    def __init__(self, vocab_size: int, embed_dimension: int = EMBED_DIMENSION):
        super(SkipGramModel, self).__init__()
        self.embeddings = nn.Embedding(
            num_embeddings=vocab_size,
            embedding_dim=embed_dimension,
            max_norm=EMBED_MAX_NORM,
        )
        self.linear = nn.Linear(
            in_features=embed_dimension,
            out_features=vocab_size,
        )

    def forward(self, inputs_):
        x = self.embeddings(inputs_)
        x = self.linear(x)
        return x


vocab_size = len(words)  # Change to your vocab length(train+test)
model = SkipGramModel(vocab_size)
