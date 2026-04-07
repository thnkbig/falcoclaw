"""HTTP client for FalcoClaw — coming in v0.2.0."""
import httpx
from typing import Optional

class FalcoClawClient:
    def __init__(self, endpoint: str = "http://localhost:2804", api_key: Optional[str] = None):
        self.endpoint = endpoint
        self.api_key = api_key

    def query_alerts(self, **kwargs):
        """Query FalcoClaw alert history."
        raise NotImplementedError("Coming in v0.2.0")

    def trigger_action(self, **kwargs):
        """Trigger a FalcoClaw response action."
        raise NotImplementedError("Coming in v0.2.0")
