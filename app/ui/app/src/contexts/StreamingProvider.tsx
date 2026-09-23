import {
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { DownloadEvent } from "@/gotypes";
import { StreamingContext } from "./streaming-context";

export function StreamingProvider({ children }: { children: ReactNode }) {
  const [streamingChatIds, setStreamingChatIds] = useState<Set<string>>(
    new Set(),
  );
  const [loadingChats, setLoadingChats] = useState<Set<string>>(new Set());
  const [abortControllers, setAbortControllers] = useState<
    Map<string, AbortController>
  >(new Map());
  const [downloadProgress, setDownloadProgress] = useState<
    Map<string, DownloadEvent>
  >(new Map());

  const contextValue = useMemo(
    () => ({
      streamingChatIds,
      setStreamingChatIds,
      loadingChats,
      setLoadingChats,
      abortControllers,
      setAbortControllers,
      downloadProgress,
      setDownloadProgress,
    }),
    [streamingChatIds, loadingChats, abortControllers, downloadProgress],
  );

  return (
    <StreamingContext.Provider value={contextValue}>
      {children}
    </StreamingContext.Provider>
  );
}
