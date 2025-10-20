(defcolumns (A :i16) (B :i16) (C :i16))
;; returns 1 if A==0 or A+1==0; otherwise, 0.
(defun (isz-A-or-Am1) (- 1 (~ (* A (- A 1)))))
;; returns non-zero value if A==0
(defun (isz-A) (* (+ 1 A) (isz-A-or-Am1)))
;; returns non-zero value if A+1==0
(defun (isz-Am1) (* A (isz-A-or-Am1)))
;; A==0 ==> B==0
(defconstraint c1 () (== 0 (* (isz-A) B)))
;; A+1==0 ==> B==0
(defconstraint c2 () (== 0 (* (isz-Am1) B)))
;; A!=0 && A+1!=0 ==> C==0
(defconstraint c3 () (== 0 (* A (- A 1) C)))
