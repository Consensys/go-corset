(defpurefun (vanishes! x) (== 0 x))
(defcolumns (A :i2@prove) (B :i16) (C :i16))

;; returns non-zero value if A is zero
(defun (isz-A) (* (- A 1) (- A 2) (- A 3)))

(defconstraint c1 () (vanishes! (* (isz-A) B)))
(defconstraint c2 () (vanishes! (* A C)))
