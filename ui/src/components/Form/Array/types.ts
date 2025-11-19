export type RenderItem<T> = ({
  value,
  setValue,
}: {
  value: T;
  setValue: (value: T) => void;
}) => JSX.Element;
