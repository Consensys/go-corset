;;error:5:27-29:unknown symbol
;;error:5:24-32:expected loobean constraint (found 𝔽)
(defcolumns A)
(defun (id x) x)
(defconstraint test () (+ id A))
