import {
  Dispatch,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
  useState,
} from "react";

type FormValidationContextType = {
  missingFields: string[];
  setMissingFields: Dispatch<SetStateAction<string[]>>;
};

const FormValidationContext = createContext<FormValidationContextType | null>(
  null
);

export const FormValidationContextProvider = ({
  children,
}: PropsWithChildren) => {
  const [missingFields, setMissingFields] = useState<string[]>([]);
  return (
    <FormValidationContext.Provider value={{ missingFields, setMissingFields }}>
      {children}
    </FormValidationContext.Provider>
  );
};

export const useFormValidationContext = () => {
  const context = useContext(FormValidationContext);

  if (!context)
    throw new Error(
      "useFormValidationContext must be used within FormValidationContextProvider"
    );

  const formValidation = context;
  return formValidation;
};
