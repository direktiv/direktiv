import {
  Dispatch,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
  useState,
} from "react";

type RequiredFieldsContextType = {
  missingFields: string[];
  setMissingFields: Dispatch<SetStateAction<string[]>>;
};

const RequiredFieldsContext = createContext<RequiredFieldsContextType | null>(
  null
);

export const RequiredFieldsContextProvider = ({
  children,
}: PropsWithChildren) => {
  const [missingFields, setMissingFields] = useState<string[]>([]);
  return (
    <RequiredFieldsContext.Provider value={{ missingFields, setMissingFields }}>
      {children}
    </RequiredFieldsContext.Provider>
  );
};

export const useRequiredFieldsContext = () => {
  const context = useContext(RequiredFieldsContext);

  if (!context)
    throw new Error(
      "useRequiredFieldsContext must be used within RequiredFieldsContextProvider"
    );

  const requiredFields = context;
  return requiredFields;
};
