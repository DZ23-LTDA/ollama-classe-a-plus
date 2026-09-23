import { createContext, type Dispatch, type SetStateAction } from "react";
import { DownloadEvent } from "@/gotypes";

export interface StreamingContextType {
  streamingChatIds: Set<string>;
  setStreamingChatIds: Dispatch<SetStateAction<Set<string>>>;
  loadingChats: Set<string>;
  setLoadingChats: Dispatch<SetStateAction<Set<string>>>;
  abortControllers: Map<string, AbortController>;
  setAbortControllers: Dispatch<SetStateAction<Map<string, AbortController>>>;
  downloadProgress: Map<string, DownloadEvent>;
  setDownloadProgress: Dispatch<SetStateAction<Map<string, DownloadEvent>>>;
}

export const StreamingContext = createContext<
  StreamingContextType | undefined
>(undefined);
