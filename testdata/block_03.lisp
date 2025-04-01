(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint c1 ()
  (if (== X 0)
      ;; if X==0 then Y == Z
      (begin
       (vanishes! Y)
       (vanishes! Z))
      ;; else X == Y && (Y == 0 || Z == 0)
      (begin
       (vanishes! (- X Y))
       (vanishes! (* Y Z)))))
;; Z is always 0!
(defproperty a1 (== 0 Z))
