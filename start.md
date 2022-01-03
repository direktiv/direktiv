# Yarn / Node

Must be version v15.14.0 or earlier

```
sudo npm install --global yarn
```

For development it needs [*CORS Everywhere*](https://addons.mozilla.org/en-US/firefox/addon/cors-everywhere/?utm_source=addons.mozilla.org&utm_medium=referral&utm_content=search) and can be started like this:

```
DIREKTIV_API=http://localhost:8080/api/ yarn start
```
  
# Docker container

Running backend container:

```
docker run  --privileged -p 8080:80 -ti gerke74/direktiv-kube:ui
```

# Figma layout link

https://www.figma.com/file/KZj2mFKlK5BWxmO7zMV8OD/Direktiv?node-id=0%3A1


# Icons

https://tablericons.com/


# API

https://docs.direktiv.io/v0.5.10/api/

or in dev version:

https://github.com/direktiv/direktiv.github.io/tree/v0.6.0