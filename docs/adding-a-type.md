# Adding A Media Type

To add a media type go to `types/types.go` and find the section declaring all `TY_*` variables.

Add `TY_<your_type> MediTypes = "<your_type>"` to the list, be sure that the string version is PascalCase.

Next, add your type to the ListMediaTypes function.
