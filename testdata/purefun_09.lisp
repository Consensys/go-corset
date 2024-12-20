(defun (id (x :binary)) x)
(defun (f x) (id x))
(defcolumns X (Y :binary))
(defconstraint c1 () (- X (f Y)))
