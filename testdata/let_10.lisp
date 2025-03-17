(defpurefun (f x) (let ((xp1 (+ x 1))) xp1))

(defcolumns (A :i16) (B :i16))
(defconstraint c1 ()
  (if (== 0 A)
      (== 0 (f (- B 1)))))
