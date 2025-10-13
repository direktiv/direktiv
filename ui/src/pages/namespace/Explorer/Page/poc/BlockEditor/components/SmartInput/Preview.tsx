import { Fragment } from "react/jsx-runtime";
import { localVariableNamespace } from "../../../schema/primitives/variable";
import { useTranslation } from "react-i18next";

export const Preview = ({ path }: { path: string[] }) => {
  const { t } = useTranslation();

  const isLocalVarNamespace = path[0] === localVariableNamespace;

  const placeholders = [
    t("direktivPage.blockEditor.smartInput.templatePlaceholders.namespace"),
    t("direktivPage.blockEditor.smartInput.templatePlaceholders.id"),
    ...(isLocalVarNamespace
      ? []
      : [
          t("direktivPage.blockEditor.smartInput.templatePlaceholders.pointer"),
        ]),
  ];

  const previewLength = Math.max(placeholders.length, path.length);

  return (
    <>
      {"{{"}
      {Array.from({ length: previewLength }, (_, index) => (
        <Fragment key={index}>
          {path[index] ? (
            <span className="text-gray-12 dark:text-gray-8">{path[index]}</span>
          ) : (
            <span className="italic text-gray-10">{placeholders[index]}</span>
          )}
          {index < previewLength - 1 && <span className="text-gray-10">.</span>}
        </Fragment>
      ))}
      {"}}"}
    </>
  );
};
