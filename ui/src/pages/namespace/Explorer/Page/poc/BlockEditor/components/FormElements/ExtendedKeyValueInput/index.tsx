import { ArrayForm } from "~/components/Form/Array";
import { Card } from "~/design/Card";
import { ExtendedKeyValueType } from "../../../../schema/primitives/extendedKeyValue";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { PropsWithChildren } from "react";
import { ValueInput } from "./ValueInput";
import { useTranslation } from "react-i18next";

type ExtendedKeyValueInputProps = {
  label: string;
  field: {
    value: ExtendedKeyValueType[] | undefined;
    onChange: (value: ExtendedKeyValueType[]) => void;
  };
  smart?: boolean;
};

export const ExtendedKeyValueInput = ({
  field,
  label,
  smart = false,
}: ExtendedKeyValueInputProps) => {
  const { t } = useTranslation();
  return (
    <Fieldset label={label}>
      <div className="flex flex-col gap-y-5">
        <ArrayForm
          value={field.value || []}
          onChange={field.onChange}
          emptyItem={{
            key: "",
            value: {
              type: "string",
              value: "",
            },
          }}
          wrapItem={(children) => (
            <Card className="flex flex-col gap-3 p-3" noShadow>
              {children}
            </Card>
          )}
          renderItem={({ value: itemValue, setValue }) => (
            <>
              <Container>
                <Label>
                  {t("direktivPage.blockEditor.blockForms.keyValue.key")}
                </Label>
                <Input
                  className="w-full"
                  placeholder={t(
                    "direktivPage.blockEditor.blockForms.keyValue.key"
                  )}
                  value={itemValue.key}
                  onChange={(e) => {
                    setValue({
                      ...itemValue,
                      key: e.target.value,
                    });
                  }}
                />
              </Container>
              <Container>
                <Label>
                  {t("direktivPage.blockEditor.blockForms.keyValue.value")}
                </Label>
                <div className="flex w-full items-center gap-2">
                  <ValueInput
                    smart={smart}
                    value={itemValue.value}
                    onChange={(newValue) => {
                      setValue({
                        ...itemValue,
                        value: newValue,
                      });
                    }}
                  />
                </div>
              </Container>
            </>
          )}
        />
      </div>
    </Fieldset>
  );
};

const Label = ({ children }: PropsWithChildren) => (
  <label className="w-[55px] grow text-sm">{children}</label>
);

const Container = ({ children }: PropsWithChildren) => (
  <div className="flex items-center gap-2">{children}</div>
);
