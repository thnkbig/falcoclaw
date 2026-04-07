"""Hermes plugin registration — coming in v0.2.0."""
from .client import FalcoClawClient

def register_plugin(hermes):
    """Register the FalcoClaw plugin with Hermes."
    hermes.register("falcoclaw", FalcoClawClient)
