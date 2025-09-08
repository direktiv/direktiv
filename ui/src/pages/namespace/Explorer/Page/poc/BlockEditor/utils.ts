import { DirektivPagesType } from "../schema";
import { QueryType } from "../schema/procedures/query";
import { keyValueArrayToObject } from "../PageCompiler/primitives/keyValue/utils";
import { useRef } from "react";

export const clonePage = (page: DirektivPagesType): DirektivPagesType =>
  structuredClone(page);

export const queryToUrl = (query: QueryType) => {
  let { url } = query;

  const searchParams = new URLSearchParams(
    keyValueArrayToObject(query.queryParams ?? [])
  );

  const paramsString = searchParams.toString();

  if (paramsString) {
    url = url.concat("?", paramsString);
  }

  return url;
};

export const useInsertText = <
  T extends HTMLTextAreaElement | HTMLInputElement,
>() => {
  const ref = useRef<T>(null);

  const insertText = (snippet: string) => {
    const element = ref.current;
    if (!element) return;

    const start = element.selectionStart ?? 0;
    const end = element.selectionEnd ?? 0;
    const value = element.value;

    const newValue = value.slice(0, start) + snippet + value.slice(end);

    element.value = newValue;

    // Move cursor after the inserted text
    const cursorPos = start + snippet.length;
    element.selectionStart = element.selectionEnd = cursorPos;

    // Trigger an input event so React state (if controlled) can update
    element.dispatchEvent(new Event("input", { bubbles: true }));
  };

  return { ref, insertText };
};
