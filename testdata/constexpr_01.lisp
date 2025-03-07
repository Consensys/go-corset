(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16) (Y :i16))
;; X == Y + n - n
(defconstraint c1 ()
  (vanishes! (- X Y (+ 1 1) (- 0 2))))
(defconstraint c2 ()
  (vanishes! (- X Y (+ 1 1 1) (- 0 1 2))))
(defconstraint c3 ()
  (vanishes! (- X Y (+ 2 1) (- 0 2 1))))
