import { Dispatch, KeyboardEvent, SetStateAction } from "react";

export type RenderItem<T> = ({
  value,
  setValue,
  onChange,
  handleKeyDown,
}: {
  value: T;
  setValue: Dispatch<SetStateAction<T>>;
  onChange: (item: T) => void;
  handleKeyDown: (event: KeyboardEvent<HTMLInputElement>) => void;
}) => JSX.Element;

export type IsValidItem<T> = (item?: T) => boolean;
