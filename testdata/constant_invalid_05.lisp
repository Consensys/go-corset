;;error:3:20-21:not permitted in pure context
(defcolumns (X :i16))
(defconst ONE (+ 1 X))
