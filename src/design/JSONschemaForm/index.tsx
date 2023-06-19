import {
  ArrayFieldTemplateItemType,
  ArrayFieldTemplateProps,
  BaseInputTemplateProps,
  DescriptionFieldProps,
  RJSFSchema,
  RegistryWidgetsType,
  TitleFieldProps,
  WidgetProps,
} from "@rjsf/utils";
import {
  ChevronDownIcon,
  ChevronUpIcon,
  MinusIcon,
  PlusIcon,
} from "lucide-react";
import React, { ComponentProps, useMemo } from "react";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";

import Button from "../Button";
import { Checkbox } from "../Checkbox";
import Form from "@rjsf/core";
import Input from "../Input";
import validator from "@rjsf/validator-ajv8";

const CustomSelectWidget: React.FC<WidgetProps> = (props) => (
  <div className="my-4">
    <Select onValueChange={props.onChange}>
      <SelectTrigger value={props.value}>
        <SelectValue
          placeholder={props.value ? props.value : `Select ${props.label}`}
        >
          {/* 
          the blank space is weirdly important here, otherwise the first change
          of this select will result in the select showing an empty text.
           */}
          {props.value}{" "}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {props.options.enumOptions?.map((op) => (
            <SelectItem key={`select-${op.value}`} value={op.value}>
              {op.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
);

const SubmitButton = () => null;

const ArrayFieldTemplateItem = (props: ArrayFieldTemplateItemType) => (
  <div className="flex flex-row items-end gap-4">
    <div className="grow">{props.children}</div>
    <div className="mb-2 flex w-min flex-row gap-2">
      <Button
        disabled={!props.hasMoveDown}
        onClick={(e) => {
          props.onReorderClick(props.index, props.index + 1)(e);
        }}
        icon
      >
        <ChevronDownIcon />
      </Button>
      <Button
        disabled={!props.hasMoveUp}
        onClick={(e) => {
          props.onReorderClick(props.index, props.index - 1)(e);
        }}
        icon
      >
        <ChevronUpIcon />
      </Button>
      <Button
        variant="destructive"
        onClick={(e) => {
          props.onDropIndexClick(props.index)(e);
        }}
        icon
      >
        <MinusIcon />
      </Button>
    </div>
  </div>
);

const ArrayFieldTemplate = (props: ArrayFieldTemplateProps) => (
  <div>
    {props.items.map((element) => (
      <ArrayFieldTemplateItem {...element} key={`array-item-${element.key}`} />
    ))}
    {props.canAdd && (
      <div className="inline-block w-full divide-y divide-solid">
        <Button onClick={props.onAddClick} icon className="float-right mt-4">
          <PlusIcon />
        </Button>
        <div className="mt-2 w-full" />
      </div>
    )}
  </div>
);

const BaseInputTemplate = (props: BaseInputTemplateProps) => {
  const type = useMemo(() => {
    if (props.schema.type === "integer") {
      return "number";
    } else if (props.type === "file") {
      return "file";
    } else {
      return undefined;
    }
  }, [props.schema.type, props.type]);

  return (
    <Input
      defaultValue={props.value}
      className="mb-2 mt-1 w-full"
      type={type}
      required={props.required}
      id={props.id}
      onChange={(e) => {
        if (type === "file") {
          if (e.target.files && e.target.files.length > 0) {
            const reader = new FileReader();
            reader.onloadend = () => {
              props.onChange(reader.result);
            };
            reader.readAsDataURL(e.target.files[0] as Blob);
          } else {
            props.onChange(e.target.files);
          }
        } else {
          props.onChange(e.target.value);
        }
      }}
    />
  );
};

const TitleFieldTemplate = (props: TitleFieldProps) => {
  const { id, required, title } = props;
  return (
    <header
      id={id}
      className="mb-4 font-semibold text-gray-12 dark:text-gray-dark-12"
    >
      {title}
      {required && <mark>*</mark>}
    </header>
  );
};
const DescriptionFieldTemplate = (props: DescriptionFieldProps) => {
  const { description } = props;
  return (
    <div className="mb-2 text-gray-8 dark:text-gray-dark-8">{description}</div>
  );
};

const CustomCheckbox = (props: WidgetProps) => (
  <div className="flex space-x-2 p-2 ">
    <Checkbox
      onClick={() => props.onChange(!props.value)}
      id={`wgt-checkbox-${props.id}`}
    />
    <div className="grid gap-1.5 leading-none">
      <label
        htmlFor={`wgt-checkbox-${props.id}`}
        className="text-sm font-medium leading-none text-gray-10 peer-disabled:cursor-not-allowed peer-disabled:opacity-70 dark:text-gray-dark-10"
      >
        {props.label}
      </label>
    </div>
  </div>
);

const widgets: RegistryWidgetsType = {
  CheckboxWidget: CustomCheckbox,
  SelectWidget: CustomSelectWidget,
};

type JSONSchemaFormProps = Omit<
  // copy the props from the original form, and remove the ones we want to implement ourselves
  ComponentProps<typeof Form>,
  "schema" | "templates" | "validator" | "widgets"
> & {
  schema: RJSFSchema;
};

export const JSONSchemaForm: React.FC<JSONSchemaFormProps> = ({
  schema,
  ...props
}) => (
  <Form
    schema={schema}
    templates={{
      BaseInputTemplate,
      TitleFieldTemplate,
      ArrayFieldTemplate,
      DescriptionFieldTemplate,
      ButtonTemplates: {
        SubmitButton,
      },
    }}
    validator={validator}
    widgets={widgets}
    {...props}
  />
);

JSONSchemaForm.displayName = "JSONSchemaForm";
