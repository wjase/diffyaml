# diffyaml

## Semantic difference for yaml files

diffyam compares two yaml files and produces a list of changes which reflect a
knowledge of the structure of yaml files. Yaml has three main types of nodes:
   * Scalar values (strings, numbers)
   * Mapped values (name=john, age=25)
   * Sequence values - an ordered list of items

Of course all of these can be mixed and matched in any combo.

yaml diff makes assumptions when comparing each node in the tree based on its type.

## Output

The report produced by diffyam is itself a yaml sequence of ChangeLog entries which can have
   * Path - A path to the affected node. Adds refer to the item in the new yaml file. Deletions refer
     to the old yaml file.
   * To, ToIndex - The yaml.Node in the new document for Adds,Moves
   * From, FromIndex - The yaml.Node in the original document for Deletes, Moves

## Usage Syntax - command line

    diffyaml filePath1 filePath2

Will produce the change log to stdout.


## Usage - golang library

You can use the Changelog items produced by diffyam to do more specific analysis relevent to a given domain, eg a kubernetes resource defintion or openapi spec.
