// src/i18n/config.ts

// Core i18next library.
import i18n from "i18next";                      
// Bindings for React: allow components to
// re-render when language changes.
import { initReactI18next } from "react-i18next";
import LanguageDetector from "i18next-browser-languagedetector";

i18n
  // Add React bindings as a plugin.
  .use(initReactI18next)
  .use(LanguageDetector)
  // Initialize the i18next instance.
  .init({
    // Fallback locale used when a translation is
    // missing in the active locale. Again, use your
    // preferred locale here. 
    fallbackLng: "en",

    // Enables useful output in the browser’s
    // dev console.
    debug: true,

    // Normally, we want `escapeValue: true` as it
    // ensures that i18next escapes any code in
    // translation messages, safeguarding against
    // XSS (cross-site scripting) attacks. However,
    // React does this escaping itself, so we turn 
    // it off in i18next.
    interpolation: {
      escapeValue: false,
    },

    // Translation messages. Add any languages
    // you want here.
    resources: {
      // English
      en: {
        // `translation` is the default namespace.
        // More details about namespaces shortly.
        translation: {
          // Generic
          copy_url: "Copy URL",
          copied: "Copied!",
          download_button: "Download",

          // Share page
          drag_area: "Drag files here or click to select files",

          sorry_share_expired: "Sorry, this share has expired.",
          share_does_not_exists: "Share does not exists.",
          please_check_link: "Please check the link used to access this page.",
          reload: "Reload",

          download_all: "Download all",

          // Shares page
          no_shares: "There are currently no shares.",
          create_share: "Create Share",
          your_shares: "Your Shares",
          other_shares: "Other Shares",
          create: "Create",

          // Share component
          guests_can_upload: "Guests can upload",
          guests_can_download: "Guests can download",
          guests_can_upload_and_download: "Guests can upload & download",
          item: "file",
          items: "files",
          empty: "empty",
          day_left: "day left",
          days_left: "days left",
          unlimited: "Unlimited",
          expired: "Expired",
          created: "Created",
          delete_share: "Delete Share",
          edit_share: "Edit",
          update: "Update",

          // Item Component
          delete_file: "Delete",
          delete_this_item: "Delete this file?",

          // Share editor
          exposure: "Exposure",
          you_want_to: "You want to :",
          send: "Send",
          receive: "Receive",
          both: "Both",
          renew_share: "Renew validity",

          validity: "Validity",
          number_of_days_the_share_is_valid: "Number of days the share is valid. 0 is unlimited.",
          description: "Description",

          // Markdown Editor
          message: "Message",
          markdown_description: "This markdown will be displayed to the user",

          // Haffix
          shares: "Shares",
          logout: "Logout",
        },
      },
      // French
      fr: {
        translation: {
            // Generic
            copy_url: "Copier le lien",
            copied: "Copié!",
            download_button: "Télécharger",

            // Share page
            drag_area: "Glissez des fichiers ou cliquez pour sélectionner",

            sorry_share_expired: "Désolé, ce partage a expiré.",
            share_does_not_exists: "Ce partage n'existe pas.",
            please_check_link: "Merci de vérifier le lien qui vous a été transmis.",
            reload: "Recharger",

            download_all: "Tout Télécharger",

            // Shares page
            no_shares: "Il n'y a aucun partage.",
            create_share: "Créer un partage",
            your_shares: "Vos Partages",
            other_shares: "Autres Partages",
            create: "Créer",

            // Share component
            guests_can_upload: "Les invités peuvent envoyer",
            guests_can_download: "Les invités peuvent télécharger",
            guests_can_upload_and_download: "Les invités peuvent envoyer & télécharger",
            item: "fichier",
            items: "fichiers",
            empty: "vide",
            day_left: "jour restant",
            days_left: "jours restant",
            unlimited: "Illimité",
            expired: "Expiré",
            created: "Créé le",
            delete_share: "Supprimer",
            edit_share: "Modifier",
            update: "Mettre à jour",

            // Item Component
            delete_file: "Supprimer",
            delete_this_item: "Supprimer ce fichier ?",

            // Share editor
            exposure: "Type de partage",
            you_want_to: "Vous souhaitez :",
            send: "Envoyer",
            receive: "Reçevoir",
            both: "Les deux",
            renew_share: "Étendre la validité",

            validity: "Expiration",
            number_of_days_the_share_is_valid: "Nombre de jours pendant lesquels le partage est valide. 0 signifie illimité.",
            description: "Description",

            // Markdown Editor
            message: "Message",
            markdown_description: "Ce markdown sera affiché à l'utilisateur",

            // Haffix
            shares: "Partages",
            logout: "Quitter",
        },
      },
      // German
      de: {
        // `translation` is the default namespace.
        // More details about namespaces shortly.
        translation: {
          // Generic
          copy_url: "URL kopieren",
          copied: "Kopiert!",
          download_button: "Download",

          // Share page
          drag_area: "Dateien hierhin ziehen oder auswählen",

          sorry_share_expired: "Entschuldigung, diese Freigabe ist abgelaufen.",
          share_does_not_exists: "Freigabe existiert nicht.",
          please_check_link: "Bitte überprüfen Sie den für diesen Aufruf verwendeten Link auf Fehler.",
          reload: "Aktualisieren",

          download_all: "Alles downloaden",

          // Shares page
          no_shares: "Es sind derzeit keine Freigaben vorhanden.",
          create_share: "Freigabe erstellen",
          your_shares: "Ihre Freigaben",
          other_shares: "Andere Freigaben",
          create: "Erstellen",

          // Share component
          guests_can_upload: "Empfänger darf uploaden",
          guests_can_download: "Empfänger darf downloaden",
          guests_can_upload_and_download: "Empfänger darf uploaden & downloaden",
          item: "Datei",
          items: "Dateien",
          empty: "leer",
          day_left: "Tag übrig",
          days_left: "Tage übrig",
          unlimited: "Unbegrenzt",
          expired: "Abgelaufen",
          created: "Erstellt",
          delete_share: "Freigabe löschen",
          edit_share: "Bearbeiten",
          update: "Aktualisieren",

          // Item Component
          delete_file: "Löschen",
          delete_this_item: "Diese Datei löschen?",

          // Share editor
          exposure: "Art der Freigabe",
          you_want_to: "Sie wollen Dateien :",
          send: "Versenden",
          receive: "Empfangen",
          both: "Beides",

          validity: "Dauer der Gültigkeit",
          number_of_days_the_share_is_valid: "Anzahl der Tage bis die Freigabe abläuft. 0 bedeutet unbegrenzt.",
          description: "Beschreibung",

          // Markdown Editor
          message: "Nachricht",
          markdown_description: "Diese Markdown-Nachricht wird dem Empfänger angezeigt",

          // Haffix
          shares: "Freigaben",
          logout: "Abmelden",
        },
      },
    },
  });

export default i18n;