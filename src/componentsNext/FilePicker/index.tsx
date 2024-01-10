import { ArrowLeftToLineIcon, FolderUp } from "lucide-react";
import {
  Filepicker,
  FilepickerHeading,
  FilepickerList,
  FilepickerListItem,
  FilepickerMessage,
  FilepickerSeparator,
} from "~/design/Filepicker";
import { Fragment, useState } from "react";

import { ButtonBar } from "~/design/ButtonBar";
import { FilePathSegments } from "./FilepathSegments";
import { FilepickerItem } from "./FilepickerItem";
import Input from "~/design/Input";
import { NodeSchemaType } from "~/api/tree/schema/node";
import { analyzePath } from "~/util/router/utils";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const convertFileToPath = (string?: string) =>
  analyzePath(string).parent?.absolute ?? "/";

const FilePicker = ({
  namespace,
  defaultPath,
  onChange,
  selectable,
}: {
  namespace?: string;
  defaultPath?: string;
  onChange?: (filePath: string) => void;
  selectable?: (node: NodeSchemaType) => boolean;
}) => {
  const [path, setPath] = useState(convertFileToPath(defaultPath));
  const [inputValue, setInputValue] = useState(defaultPath ? defaultPath : "");

  const { data, isError } = useNodeContent({
    path,
    namespace,
  });

  const { t } = useTranslation();

  const { parent, isRoot, segments } = analyzePath(path);

  const results = data?.children?.results ?? [];
  const noResults = data?.children?.results.length ? false : true;

  const pathNotFound = isError;
  const emptyDirectory = !isError && noResults;
  const folderUpButton = !isError && !noResults && !isRoot;

  return (
    <ButtonBar>
      <Filepicker
        buttonText={t("components.filepicker.buttonText")}
        onClick={() => {
          setPath(convertFileToPath(inputValue));
        }}
        className="w-44"
      >
        <FilepickerHeading>
          <FilePathSegments
            segments={segments}
            setPath={(path) => setPath(path)}
          />
        </FilepickerHeading>
        <FilepickerSeparator />
        {folderUpButton && (
          <>
            <div
              onClick={() => {
                parent ? setPath(parent.absolute) : null;
              }}
              className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
            >
              <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
            </div>
            <FilepickerSeparator />
          </>
        )}
        {emptyDirectory && (
          <FilepickerMessage>
            {t("components.filepicker.emptyDirectory.title", { path })}
          </FilepickerMessage>
        )}
        {pathNotFound && (
          <>
            <div
              onClick={() => {
                setPath("/");
              }}
              className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
            >
              <FilepickerListItem icon={ArrowLeftToLineIcon}>
                {t("components.filepicker.error.linkText")}
              </FilepickerListItem>
            </div>
            <FilepickerSeparator />
            <FilepickerMessage>
              {t("components.filepicker.error.title", { path })}
            </FilepickerMessage>
          </>
        )}
        {results && (
          <FilepickerList>
            <FilepickerItem
              results={results}
              selectable={selectable}
              setPath={(path) => setPath(path)}
              setInputValue={(value) => setInputValue(value)}
              onChange={(path) => onChange?.(path)}
            />
            <FilepickerSeparator />
          </FilepickerList>
        )}
      </Filepicker>
      <Input
        placeholder={t("components.filepicker.placeholder")}
        value={inputValue}
        onChange={(e) => {
          setInputValue(e.target.value);
          onChange?.(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export default FilePicker;
