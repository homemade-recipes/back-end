function recipe(lang, r, fav) {
	const by = lang == "en" ? "By" : "Por";
	const on = lang == "en" ? "In" : "Em";
	const heart = fav === true ? "❤️" : "🤍";
	return `<li class="list">
		<div>
			<img 
				class="lazyload"
				loading="lazy"
				data-src="/images/icon.png"
				src="${r.Picture}" 
				alt="${r.Title}"
			/>
			<a href="recipe.html?title=${r.Title}&lang=${lang}">${
				r.Title
			}</a>
			<div>
				${by}: ${r.Author}
			</div>
			<div>
				${on}: ${r.Category}
			</div>
		</div>
		<button id="${r.Title}" class="fav" onclick="toggleFav('${r.Title}')">
			${heart}
		</button>
	</li>`;
}
