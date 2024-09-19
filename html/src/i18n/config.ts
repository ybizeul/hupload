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
          drag_area: "Drag files here or click to select files",
          no_shares: "There are currently no shares.",
          create_share: "Create Share",
          your_shares: "Your Shares",
          other_shares: "Other Shares",

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

          exposure: "Exposure",
          you_want_to: "You want to :",
          send: "Send",
          receive: "Receive",
          both: "Both",

          validity: "Validity",
          number_of_days_the_share_is_valid: "Number of days the share is valid. 0 is unlimited.",
          description: "Description",

          create: "Create",

          message: "Message",
          markdown_description: "This markdown will be displayed to the user",

          sorry_share_expired: "Sorry, this share has expired.",
          share_does_not_exists: "Share does not exists.",
          please_check_link: "Please check the link used to access this page.",
          reload: "Reload",

          download_button: "Download",

        },
      },
      // Arabic
      fr: {
        translation: {
            drag_area: "Glissez des fichiers ou cliquez pour sélectionner",
            no_shares: "Il n'y a aucun partage.",
            create_share: "Créer un partage",
            your_shares: "Vos Partages",
            other_shares: "Autres Partages",

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

            exposure: "Type de partage",
            you_want_to: "Vous souhaitez :",
            send: "Envoyer",
            receive: "Reçevoir",
            both: "Les deux",

            validity: "Expiration",
            number_of_days_the_share_is_valid: "Nombre de jours pendant lesquels le partage est valide. 0 signifie illimité.",
            description: "Description",

            create: "Créer",

            message: "Message",
            markdown_description: "Ce markdown sera affiché à l'utilisateur",

            sorry_share_expired: "Désolé, ce partage a expiré.",
            share_does_not_exists: "Ce partage n'existe pas.",
            please_check_link: "Merci de vérifier le lien qui vous a été transmis.",
            reload: "Recharger",

            download_button: "Télécharger",
        },
      },
    },
  });

export default i18n;