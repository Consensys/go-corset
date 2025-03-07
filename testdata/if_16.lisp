(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns (A :i2@loob@prove) B C)

;; returns non-zero value if A is zero
(defun (isz-A) (* (- A 1) (- A 2) (- A 3)))

(defconstraint c1 () (vanishes! (* (isz-A) B)))
(defconstraint c2 () (vanishes! (* A C)))
