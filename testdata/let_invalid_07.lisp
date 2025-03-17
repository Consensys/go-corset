;;error:3:33-34:not permitted in pure context
(defpurefun (vanishes! x) (== 0 x))
(defpurefun (f x) (let ((xp1 (+ A 1))) xp1))

(defcolumns (A :i16) (B :i16))
(defconstraint c1 ()
  (if A
      (vanishes! (f (- B 1)))))
