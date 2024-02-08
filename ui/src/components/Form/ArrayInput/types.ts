import { KeyboardEvent } from "react";

export type RenderItem<T> = ({
  value,
  setValue,
  handleKeyDown,
}: {
  value: T;
  setValue: (value: T) => void;
  handleKeyDown: (event: KeyboardEvent<HTMLInputElement>) => void;
}) => JSX.Element;

export type IsValidItem<T> = (item?: T) => boolean;
