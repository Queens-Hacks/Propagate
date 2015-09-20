var gulp = require('gulp');
var watchify = require('gulp-watchify')
var uglify = require('gulp-uglify');

var bundlePaths = {
        src: [
            'js/*.js'
        ],
        dest: 'build'
    }
    // Hack to enable configurable watchify watching
var watching = false
gulp.task('enable-watch-mode', function() {
        watching = true
    })
    // Browserify and copy js files
gulp.task('browserify', watchify(function(watchify) {
    return gulp.src(bundlePaths.src)
        // .pipe(uglify())
        .pipe(watchify({
            watch: watching
        })).on('error', function(err) {
            console.log(err.message);
        })
        .pipe(gulp.dest(bundlePaths.dest))
}))

gulp.task('watchify', ['enable-watch-mode', 'browserify'])
    // Rerun tasks when a file changes
gulp.task('watch', ['watchify'], function() {
        // ... other watch code ...
    })
    // The default task (called when you run `gulp` from cli)
gulp.task('default', ['browserify'])
