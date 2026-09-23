import { useContext } from "react";
import { StreamingContext } from "./streamingContext";

export function useStreamingContext() {
  const context = useContext(StreamingContext);
  if (context === undefined) {
    throw new Error(
      "useStreamingContext must be used within a StreamingProvider",
    );
  }
  return context;
}
