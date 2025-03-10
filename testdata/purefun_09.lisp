(defun (id2 (x :binary)) x)
(defun (f (x :binary)) (id2 x))
(defcolumns (X :i16@loob) (Y :binary))
(defconstraint c1 () (- X (f Y)))
