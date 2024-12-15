;;error:2:1-2:blah
(defcolumns A)
(defun (id x) x)
(defconstraint test () (id A A))
