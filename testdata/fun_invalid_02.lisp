;;error:5:24-32:incorrect number of arguments (found 2)
;;error:5:25-27:ambiguous invocation
(defcolumns A)
(defun (fd x) x)
(defconstraint test () (fd A A))
