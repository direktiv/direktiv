import { Dispatch, KeyboardEvent, SetStateAction } from "react";

export type RenderItem<T> = ({
  state,
  onChange,
  setState,
  handleKeyDown,
}: {
  state: T;
  onChange: (item: T) => void;
  setState: Dispatch<SetStateAction<T>>;
  handleKeyDown: (event: KeyboardEvent<HTMLInputElement>) => void;
}) => JSX.Element;

export type IsValidItem<T> = (item?: T) => boolean;
