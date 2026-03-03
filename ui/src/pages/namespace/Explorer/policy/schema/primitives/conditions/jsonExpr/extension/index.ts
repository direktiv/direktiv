import { JsonExprReservedKeys } from "../constants";
import { z } from "zod";

// when { decimal("10.0") } / when { context.source_ip.isInRange(ip("222.222.222.0/24")) };
export const ExtensionJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  z
    .record(z.array(jsonExprSchema))
    .refine((value) => {
      const keys = Object.keys(value);
      return keys.length === 1;
    })
    .refine((value) => {
      const [key] = Object.keys(value);
      return key !== undefined && !JsonExprReservedKeys.has(key);
    });
