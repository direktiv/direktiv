import { Dispatch, KeyboardEvent, SetStateAction } from "react";

export type RenderItemType<T> = ({
  state,
  onChange,
  setState,
  handleKeyDown,
}: {
  state: T;
  onChange: (newValue: T) => void;
  setState: Dispatch<SetStateAction<T>>;
  handleKeyDown: (event: KeyboardEvent<HTMLInputElement>) => void;
}) => JSX.Element;
