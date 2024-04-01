const btn = document.querySelector("input[type='submit']");
const btnText = btn.textContent;

btn.addEventListener("mouseover", () => {
  btn.style.background = "blue";
  btn.style.color = "white";
  btn.style.boxShadow = "0 0 10px blue";
  btn.textContent = "Activate";
});

btn.addEventListener("mouseout", () => {
  btn.style.background = "#2c3e50";
  btn.style.color = "white";
  btn.style.boxShadow = "none";
  btn.textContent = btnText;
});

const tiles = document.querySelectorAll(".tile");

for (let i = 0; i < tiles.length; i++) {
  tiles[i].style.backgroundImage = `linear-gradient(to right, rgb(255, 0, 0), rgb(0, 255, 0), rgb(0, 0, 255))`;
}
tiles.forEach((tile) => {
  tile.addEventListener("mouseover", () => {
    tile.style.background = "blue";
    tile.style.color = "white";
  });
  tile.addEventListener("mouseout", () => {
    tile.style.background = "#bdc3c7";
    tile.style.color = "black";
  });
});