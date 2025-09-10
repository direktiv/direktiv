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
