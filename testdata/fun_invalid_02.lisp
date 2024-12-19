;;error:5:24-32:incorrect number of arguments (found 2)
;;error:5:24-32:ambiguous invocation
(defcolumns A)
(defun (id x) x)
(defconstraint test () (id A A))
