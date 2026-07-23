package detect

// detectKotlin claims a Gradle project that applies a Kotlin plugin and is not
// an Android project (Android is checked first). This covers Kotlin/JVM and
// non-Android Kotlin Multiplatform.
func detectKotlin(dir string) *Project {
	if !hasGradleFiles(dir) {
		return nil
	}
	if gradleContains(dir, "com.android.application") || gradleContains(dir, "com.android.library") {
		return nil
	}

	multiplatform := gradleContains(dir, "kotlin(\"multiplatform\")") ||
		gradleContains(dir, "org.jetbrains.kotlin.multiplatform")
	jvm := gradleContains(dir, "kotlin(\"jvm\")") ||
		gradleContains(dir, "org.jetbrains.kotlin.jvm")
	if !multiplatform && !jvm {
		return nil
	}

	subtype := "jvm"
	var flags []string
	if multiplatform {
		subtype = "multiplatform"
		flags = append(flags, "kmp")
	}
	return &Project{Path: dir, Stack: Kotlin, Subtype: subtype, Confidence: 0.85, Flags: flags}
}
