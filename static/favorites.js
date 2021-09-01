function isFavorite(title) {
	let button = document.getElementById(title);

	if (localStorage.getItem(title)) {
		button.innerText = "❤️";
	} else {
		button.innerText = "🤍";
	}
}

function toggleFav(title) {
	let button = document.getElementById(title);
	if (localStorage.getItem(title)) {
		localStorage.removeItem(title);
		button.innerText = "🤍";
	} else {
		// Get current lang
		const lang = document.getElementsByTagName("html")[0].lang;
		localStorage.setItem(title, lang);
		button.innerText = "❤️";
	}
}

function getFavorites(lang) {
	let list = document.getElementById("list");
	for (let i = 0; i < localStorage.length; i++) {
		if (localStorage.getItem(localStorage.key(i)) != lang) {
			continue;
		}
		
		// Make request and show
		fetch("api?lang="+lang+"&name="+localStorage.key(i)).then(r => {
			if (r.status == 200) {
				r.json().then(j => {
					list.innerHTML += recipe(lang, j[0], true);
				});
			}
		});
	}
}
