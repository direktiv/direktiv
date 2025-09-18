import { Fragment, PropsWithChildren, ReactNode, useState } from "react";

import { ArrayItem } from "./ArrayItem";
import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Plus } from "lucide-react";
import { RenderItem } from "./types";

type ArrayFormProps = <T>(
  props: {
    defaultValue: T[];
    emptyItem: T;
    onChange: (newArray: T[]) => void;
    renderItem: RenderItem<T>;
    wrapItem?: (children: ReactNode) => JSX.Element;
  } & PropsWithChildren
) => JSX.Element;

export const ArrayForm: ArrayFormProps = ({
  children,
  defaultValue,
  emptyItem,
  onChange,
  renderItem,
  wrapItem = (children) => <ButtonBar>{children}</ButtonBar>,
}) => {
  type Item = (typeof defaultValue)[number];
  const [items, setItems] = useState(defaultValue);

  const addItem = (newItem: Item) => {
    const newValue = [...items, newItem];
    setItems(newValue);
    onChange(newValue);
  };

  const updateAtIndex = (index: number, value: Item) => {
    const newItems = items.map((oldValue, oldIndex) => {
      if (oldIndex === index) {
        return value;
      }
      return oldValue;
    });
    setItems(newItems);
    onChange(newItems);
  };

  const deleteAtIndex = (index: number) => {
    const newItems = items.filter((_, oldIndex) => oldIndex !== index);
    setItems(newItems);
    onChange(newItems);
  };

  return (
    <>
      {items?.map((item, index) => (
        <Fragment key={index}>
          {wrapItem(
            <ArrayItem
              defaultValue={item}
              renderItem={renderItem}
              onUpdate={(value) => updateAtIndex(index, value)}
              onDelete={() => deleteAtIndex(index)}
            />
          )}
        </Fragment>
      ))}

      <Button
        type="button"
        variant="outline"
        onClick={(e) => {
          e.preventDefault();
          addItem(emptyItem);
        }}
      >
        <Plus /> {children}
      </Button>
    </>
  );
};
