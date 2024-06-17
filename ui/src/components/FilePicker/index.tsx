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

import { BaseFileSchemaType } from "~/api/files/schema";
import { ButtonBar } from "~/design/ButtonBar";
import { FileList } from "./FileList";
import { FilePathSegments } from "./FilepathSegments";
import Input from "~/design/Input";
import { analyzePath } from "~/util/router/utils";
import { useFile } from "~/api/files/query/file";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const convertFileToPath = (string?: string) =>
  analyzePath(string).parent?.absolute ?? "/";

const FilePicker = ({
  namespace,
  defaultPath,
  onChange,
  selectable,
  selectableFolders,
}: {
  namespace?: string;
  defaultPath?: string;
  onChange?: (filePath: string) => void;
  selectable?: (file: BaseFileSchemaType) => boolean;
  selectableFolders?: boolean;
}) => {
  const [path, setPath] = useState(convertFileToPath(defaultPath));
  const [inputValue, setInputValue] = useState(defaultPath ? defaultPath : "");

  const { data, isError } = useFile({
    path,
    namespace,
  });

  const { t } = useTranslation();

  const { parent, isRoot, segments } = analyzePath(path);

  const results = data?.type === "directory" ? data?.children ?? [] : [];
  const noResults = results.length ? false : true;

  const pathNotFound = isError;
  const emptyDirectory = !isError && noResults;
  const folderUpButton = !isError && !noResults && !isRoot;

  const selectFolders = selectableFolders || false;

  return (
    <ButtonBar>
      <Filepicker
        buttonText={t(
          selectFolders
            ? "components.folderpicker.buttonText"
            : "components.filepicker.buttonText"
        )}
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
              files={results}
              selectable={selectable}
              setPath={(path) => setPath(path)}
              setInputValue={(value) => setInputValue(value)}
              onChange={(path) => onChange?.(path)}
              selectableFolders={selectFolders}
            />
            <FilepickerSeparator />
          </FilepickerList>
        )}
      </Filepicker>
      <Input
        placeholder={t(
          selectFolders
            ? "components.folderpicker.placeholder"
            : "components.filepicker.placeholder"
        )}
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
