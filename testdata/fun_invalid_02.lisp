;;error:4:25-27:found 2 arguments, expected 1
(defcolumns (A :i16))
(defun (fd x) x)
(defconstraint test () (fd A A))
