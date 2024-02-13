function removeElement(selector) {
  var element = document.querySelector(selector);
  if (element) {
    element.remove();
  }
}

function removeModalDialog() {
  removeElement(".modal-dialog");
}
