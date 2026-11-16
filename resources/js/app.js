// The script of the site: the framework's behaviours (theme control, copy
// buttons, the menus), Basecoat's, and the site's own.
//
// YPY, the library that boosts the navigation, is not here. It is served as its
// own file, built by its own project, so it runs exactly as it was published.
import 'arandu:ui.js';
import 'arandu:basecoat.bundle.js';
import './site.js';
// Keep theme selection before the first paint; interactions load with defer.
import 'arandu:theme.js';
