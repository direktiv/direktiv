import { Fragment, PropsWithChildren, ReactNode } from "react";

import { ArrayItem } from "./ArrayItem";
import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Plus } from "lucide-react";
import { RenderItem } from "./types";

type ArrayFormProps = <T>(
  props: {
    value: T[];
    emptyItem: T;
    onChange: (newArray: T[]) => void;
    renderItem: RenderItem<T>;
    wrapItem?: (children: ReactNode) => JSX.Element;
  } & PropsWithChildren
) => JSX.Element;

export const ArrayForm: ArrayFormProps = ({
  children,
  value,
  emptyItem,
  onChange,
  renderItem,
  wrapItem = (children) => <ButtonBar>{children}</ButtonBar>,
}) => {
  type Item = (typeof value)[number];

  const addItem = (newItem: Item) => {
    const newValue = [...value, newItem];
    onChange(newValue);
  };

  const updateAtIndex = (index: number, newValue: Item) => {
    const newItems = value.map((oldValue, oldIndex) => {
      if (oldIndex === index) {
        return newValue;
      }
      return oldValue;
    });
    onChange(newItems);
  };

  const deleteAtIndex = (index: number) => {
    const newItems = value.filter((_, oldIndex) => oldIndex !== index);
    onChange(newItems);
  };

  return (
    <>
      {value?.map((item, index) => (
        <Fragment key={index}>
          {wrapItem(
            <ArrayItem
              value={item}
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
