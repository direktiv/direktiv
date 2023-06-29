import { FC } from "react";
import { NavLink } from "react-router-dom";
import { createClassNames } from "~/design/NavigationLink";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const Navigation: FC = () => {
  const namespace = useNamespace();
  const { t } = useTranslation();
  if (!namespace) return null;

  type PagesKeys = keyof typeof pages;
  return (
    <>
      {Object.entries(pages).map(([key, item]) => {
        // we should normaly avoid using "as" at this place, because we should not tell
        // TS that we know better than it, but in this case we actually do and it can
        // simply not infer the type of key at this point.
        const typedKey = key as PagesKeys;
        if (!item.icon || !item.name) return null;
        return (
          <NavLink
            key={key}
            to={item.createHref({ namespace })}
            className={({ isActive }) => createClassNames(isActive)}
            end={false}
          >
            <item.icon aria-hidden="true" />{" "}
            {t(`components.mainMenu.${typedKey}`)}
          </NavLink>
        );
      })}
    </>
  );
};

export default Navigation;
