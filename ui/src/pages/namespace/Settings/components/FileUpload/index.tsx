import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { Loader2 } from "lucide-react";
import { parseDataUrl } from "./utils";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type FileUploadProps = {
  onChange: (file: { base64String: string; mimeType: string }) => void;
};

const FileUpload = ({ onChange }: FileUploadProps) => {
  const { t } = useTranslation();
  const [isUploading, setIsUploading] = useState(false);

  const onFileLoad = (e: ProgressEvent<FileReader>) => {
    const fileContent = e.target?.result;
    if (typeof fileContent === "string") {
      const parsedDataUrl = parseDataUrl(fileContent);
      if (parsedDataUrl) {
        onChange({
          base64String: parsedDataUrl.data,
          mimeType: parsedDataUrl.mimeType,
        });
      }
    }
    setIsUploading(false);
  };

  const onFileLoadError = () => {
    setIsUploading(false);
  };

  const onFilepickerChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const fileReader = new FileReader();
    fileReader.onload = onFileLoad;
    fileReader.onerror = onFileLoadError;

    setIsUploading(true);
    fileReader.readAsDataURL(file);
  };

  return (
    <fieldset className="flex items-center gap-5">
      <label className="w-[150px] text-right" htmlFor="file-upload">
        {t("pages.settings.variables.create.file.label")}
      </label>
      <InputWithButton>
        <Input id="file-upload" type="file" onChange={onFilepickerChange} />
        {isUploading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
      </InputWithButton>
    </fieldset>
  );
};

export default FileUpload;
