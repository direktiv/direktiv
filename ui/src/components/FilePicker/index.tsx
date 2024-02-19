import { ArrowLeftToLineIcon, FolderUp } from "lucide-react";
import {
  Filepicker,
  FilepickerButton,
  FilepickerHeading,
  FilepickerList,
  FilepickerListItem,
  FilepickerMessage,
  FilepickerSeparator,
} from "~/design/Filepicker";
import { Fragment, useState } from "react";

import { ButtonBar } from "~/design/ButtonBar";
import { FileList } from "./FileList";
import { FilePathSegments } from "./FilepathSegments";
import { FileSchemaType } from "~/api/filesTree/schema";
import Input from "~/design/Input";
import { analyzePath } from "~/util/router/utils";
import { useNode } from "~/api/filesTree/query/node";
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
  selectable?: (node: FileSchemaType) => boolean;
}) => {
  const [path, setPath] = useState(convertFileToPath(defaultPath));
  const [inputValue, setInputValue] = useState(defaultPath ? defaultPath : "");

  const { data, isError } = useNode({
    path,
    namespace,
  });

  const { t } = useTranslation();

  const { parent, isRoot, segments } = analyzePath(path);

  const results = data?.children ?? [];
  const noResults = data?.children?.length ? false : true;

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
            <FilepickerButton
              onClick={() => {
                parent ? setPath(parent.absolute) : null;
              }}
            >
              <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
            </FilepickerButton>
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
            <FilepickerButton
              onClick={() => {
                setPath("/");
              }}
            >
              <FilepickerListItem icon={ArrowLeftToLineIcon}>
                {t("components.filepicker.error.linkText")}
              </FilepickerListItem>
            </FilepickerButton>
            <FilepickerSeparator />
            <FilepickerMessage>
              {t("components.filepicker.error.title", { path })}
            </FilepickerMessage>
          </>
        )}
        {results && (
          <FilepickerList>
            <FileList
              nodes={results}
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
