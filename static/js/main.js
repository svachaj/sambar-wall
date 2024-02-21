function removeElement(selector) {
  var element = document.querySelectorAll(selector);
  if (element) {
    element.forEach((el) => {
      el.remove();
    });
  }
}

function removeModalDialog() {
  removeElement(".modal-dialog");
}

// set and save dark mode to local storage
function toggleDarkMode() {
  document.documentElement.classList.toggle("dark");
  var isDarkMode = document.documentElement.classList.contains("dark");
  localStorage.setItem("dark", isDarkMode);
}

// get dark mode from local storage
function getDarkMode() {
  var isDarkMode = localStorage.getItem("dark");
  if (isDarkMode) {
    if (isDarkMode === "true") {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }
}

// Path: static/js/main.js
document.addEventListener("DOMContentLoaded", function () {
  getDarkMode();
});
