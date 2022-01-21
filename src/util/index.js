export const Config = {
    url: process.env.REACT_APP_API ? process.env.REACT_APP_API : "/api/" 
}

function fallbackCopyTextToClipboard(text) {
    var textArea = document.createElement("textarea");
    textArea.value = text;
    
    // Avoid scrolling to bottom
    textArea.style.top = "0";
    textArea.style.left = "0";
    textArea.style.position = "fixed";
  
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
  
    try {
      var successful = document.execCommand('copy');
      var msg = successful ? 'successful' : 'unsuccessful';
      if (msg) {
        console.log(msg);
      }
    } catch (err) {
      console.error('Fallback: Oops, unable to copy', err);
    }
  
    document.body.removeChild(textArea);
  }
  export function copyTextToClipboard(text) {
    if (!navigator.clipboard) {
      fallbackCopyTextToClipboard(text);
      return;
    }
    navigator.clipboard.writeText(text).then(function() {
      console.log('Async: Copying to clipboard was successful!');
    }, function(err) {
      console.error('Async: Could not copy text: ', err);
    });
  }

export function GenerateRandomKey(prefix) {
    if (!prefix) {
        prefix = "";
    }

    return prefix + Array(16).fill().map(()=>"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".charAt(Math.random()*62)).join("")
}

const PreviewableMimeTypes = ["application/json", "application/x-sh", "text/html", "application/yaml", "text/plain"]

export function CanPreviewMimeType(mime) {

  for (let index = 0; index < PreviewableMimeTypes.length; index++) {
      const pmt = PreviewableMimeTypes[index];
      if (mime.includes(pmt)) {
          return true
      }
  }

  return false
}

const MimeTypeExtensionsMap = {
  "text/plain": "txt",
  "application/json": "json",
  "application/x-sh": "shell",
  "text/html": "html",
  "text/css": "css",
  "application/yaml": "yaml",
  "image/jpeg": "jpg",
  "image/gif": "gif",
  "image/png": "png"
}

// best effort getting file extension from mimetype
export function MimeTypeFileExtension(mime) {
  for (const [mimeType, extension] of Object.entries(MimeTypeExtensionsMap)) {
    if (mime.includes(mimeType)) {
      return extension
    }
  }
  return null
}


