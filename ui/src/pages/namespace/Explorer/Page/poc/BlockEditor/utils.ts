import { DirektivPagesType } from "../schema";
import { QueryType } from "../schema/procedures/query";
import { keyValueArrayToObject } from "../PageCompiler/primitives/keyValue/utils";

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

export const addSnippetToInputValue = ({
  element,
  snippet,
  value,
  callback,
}: {
  element: HTMLInputElement | HTMLTextAreaElement;
  snippet: string;
  value: string;
  callback: (value: string) => void;
}) => {
  const start = element.selectionStart;
  const end = element.selectionEnd;

  if (start === null || end === null) {
    return;
  }

  const newValue = value.slice(0, start) + snippet + value.slice(end);
  callback(newValue);

  const cursorPos = start + snippet.length;
  element.setSelectionRange(cursorPos, cursorPos);
  element.focus();
};
