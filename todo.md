# Remains to be implemented

## html formats
firefox SHORTCUTURL attr
ie icons

## check error handling for
os.IsNotExist 

## README
| ........................ process ........................... |
| .......... parse ... | ... run ... | ... stringify ..........|

          +--------+                     +----------+
Input ->- | Parser | ->- Syntax Tree ->- | Compiler | ->- Output
          +--------+          |          +----------+
                              X
                              |
                       +--------------+
                       | Transformers |
                       +--------------+

# should not use dir name as map keys